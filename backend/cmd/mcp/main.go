package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/config"
	financeRepo "github.com/kinkando/personal-dashboard/internal/finance/repository"
	financeSvc "github.com/kinkando/personal-dashboard/internal/finance/service"
	kanbanRepo "github.com/kinkando/personal-dashboard/internal/kanban/repository"
	userRepo "github.com/kinkando/personal-dashboard/internal/user/repository"
	"github.com/kinkando/personal-dashboard/pkg/mongo"
	"github.com/kinkando/personal-dashboard/pkg/postgres"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// All log output goes to stderr — stdout is reserved for the MCP stdio protocol.
	logger := log.New(os.Stderr, "[mcp] ", log.LstdFlags)

	cfg := config.Load()

	if cfg.MCPUserFirebaseUID == "" {
		logger.Fatal("MCP_USER_FIREBASE_UID is required; set it to your Firebase UID " +
			"(sign into the web app first to provision the users row)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pgDB, err := postgres.New(ctx, cfg.PostgresDSN)
	if err != nil {
		logger.Fatalf("postgres init: %v", err)
	}
	defer pgDB.Close()

	mongoDB, err := mongo.New(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		logger.Fatalf("mongo init: %v", err)
	}
	defer func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutCancel()
		if err := mongoDB.Client().Disconnect(shutCtx); err != nil {
			logger.Printf("mongo disconnect: %v", err)
		}
	}()

	// Wire repositories and services.
	usrRepo := userRepo.New(pgDB.SQL())
	finRepo := financeRepo.New(pgDB.SQL())
	finSvc := financeSvc.New(finRepo)
	kanRepo := kanbanRepo.New(mongoDB)

	// Resolve the internal UUID for the configured Firebase UID once at startup.
	// Finance operations require a UUID; kanban uses the raw Firebase UID string.
	userUUID, err := usrRepo.GetIDByFirebaseUID(ctx, cfg.MCPUserFirebaseUID)
	if err != nil {
		logger.Fatalf("could not resolve MCP_USER_FIREBASE_UID %q: %v\n"+
			"Make sure the user has signed into the web app at least once.", cfg.MCPUserFirebaseUID, err)
	}
	if userUUID == (uuid.UUID{}) {
		logger.Fatalf("user %q not found in the users table; sign in via the web app first", cfg.MCPUserFirebaseUID)
	}

	logger.Printf("starting MCP server (user=%s)", cfg.MCPUserFirebaseUID)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "kinkando-dashboard",
		Version: "0.1.0",
	}, nil)

	registerTools(server, deps{
		finSvc:         finSvc,
		kanRepo:        kanRepo,
		userUUID:       userUUID,
		firebaseUID:    cfg.MCPUserFirebaseUID,
	})

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "[mcp] server error: %v\n", err)
		os.Exit(1)
	}
}
