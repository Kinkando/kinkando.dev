package main

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/kinkando/personal-dashboard/config"
	aichatHandler "github.com/kinkando/personal-dashboard/internal/aichat/handler"
	"github.com/kinkando/personal-dashboard/internal/auth"
	financeHandler "github.com/kinkando/personal-dashboard/internal/finance/handler"
	financeRepo "github.com/kinkando/personal-dashboard/internal/finance/repository"
	financeSvc "github.com/kinkando/personal-dashboard/internal/finance/service"
	"github.com/kinkando/personal-dashboard/internal/gemini"
	healthHandler "github.com/kinkando/personal-dashboard/internal/health/handler"
	healthRepo "github.com/kinkando/personal-dashboard/internal/health/repository"
	healthSvc "github.com/kinkando/personal-dashboard/internal/health/service"
	kanbanHandler "github.com/kinkando/personal-dashboard/internal/kanban/handler"
	kanbanRepo "github.com/kinkando/personal-dashboard/internal/kanban/repository"
	"github.com/kinkando/personal-dashboard/internal/line"
	lineHandler "github.com/kinkando/personal-dashboard/internal/line/handler"
	"github.com/kinkando/personal-dashboard/internal/mcpserver"
	medicineHandler "github.com/kinkando/personal-dashboard/internal/medicine/handler"
	medicineRepo "github.com/kinkando/personal-dashboard/internal/medicine/repository"
	medicineSvc "github.com/kinkando/personal-dashboard/internal/medicine/service"
	portfolioHandler "github.com/kinkando/personal-dashboard/internal/portfolio/handler"
	"github.com/kinkando/personal-dashboard/internal/quest"
	questHandler "github.com/kinkando/personal-dashboard/internal/quest/handler"
	questRepo "github.com/kinkando/personal-dashboard/internal/quest/repository"
	questSvc "github.com/kinkando/personal-dashboard/internal/quest/service"
	userHandler "github.com/kinkando/personal-dashboard/internal/user/handler"
	userRepo "github.com/kinkando/personal-dashboard/internal/user/repository"
	userSvc "github.com/kinkando/personal-dashboard/internal/user/service"
	workoutHandler "github.com/kinkando/personal-dashboard/internal/workout/handler"
	workoutRepo "github.com/kinkando/personal-dashboard/internal/workout/repository"
	workoutSvc "github.com/kinkando/personal-dashboard/internal/workout/service"
	"github.com/kinkando/personal-dashboard/pkg/event"
	"github.com/kinkando/personal-dashboard/pkg/middleware"
	"github.com/kinkando/personal-dashboard/pkg/mongo"
	"github.com/kinkando/personal-dashboard/pkg/postgres"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

func main() {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		log.Fatalf("load timezone: %v", err)
	}
	time.Local = loc

	cfg := config.Load()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("init zap: %v", err)
	}
	defer logger.Sync() //nolint:errcheck

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pgDB, err := postgres.New(ctx, cfg.PostgresDSN)
	if err != nil {
		logger.Fatal("postgres init", zap.Error(err))
	}
	defer pgDB.Close()

	mongoDB, err := mongo.New(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		logger.Fatal("mongo init", zap.Error(err))
	}
	defer func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutCancel()
		if err := mongoDB.Client().Disconnect(shutCtx); err != nil {
			logger.Error("mongo disconnect", zap.Error(err))
		}
	}()

	authMW, err := auth.NewMiddleware(ctx, cfg.FirebaseCredentials)
	if err != nil {
		logger.Fatal("firebase auth init", zap.Error(err))
	}

	// wire modules
	bus := event.New()

	usrRepo := userRepo.New(pgDB.SQL())
	usrSvc := userSvc.New(usrRepo)
	usrH := userHandler.New(usrSvc, usrRepo)

	finRepo := financeRepo.New(pgDB.SQL())
	finSvc := financeSvc.New(finRepo)
	finH := financeHandler.New(finSvc, usrRepo)

	kanRepo := kanbanRepo.New(mongoDB)
	kanH := kanbanHandler.New(kanRepo)

	heaRepo := healthRepo.New(pgDB.SQL())
	heaSvc := healthSvc.New(heaRepo, bus)
	heaH := healthHandler.New(heaSvc, usrRepo)

	wkRepo := workoutRepo.New(pgDB.SQL())
	wkSvc := workoutSvc.New(wkRepo, bus)
	wkH := workoutHandler.New(wkSvc, usrRepo)

	medRepo := medicineRepo.New(pgDB.SQL())
	medSvc := medicineSvc.New(medRepo, bus)
	medH := medicineHandler.New(medSvc, usrRepo)

	qstRepo := questRepo.New(pgDB.SQL())
	qstSvc := questSvc.New(qstRepo)
	qstH := questHandler.New(qstSvc, usrRepo)

	// Subscribe quest to domain events — main.go is the only place that knows
	// both producers and subscribers; neither side imports the other.
	bus.Subscribe(event.MedicineTaken, func(ctx context.Context, e event.Event) {
		_ = qstSvc.HandleSourceEvent(ctx, e.UserID, string(quest.SourceTypeMedicine))
	})
	bus.Subscribe(event.SupplementTaken, func(ctx context.Context, e event.Event) {
		_ = qstSvc.HandleSourceEvent(ctx, e.UserID, string(quest.SourceTypeSupplement))
	})
	bus.Subscribe(event.WorkoutSessionFinished, func(ctx context.Context, e event.Event) {
		_ = qstSvc.HandleSourceEvent(ctx, e.UserID, string(quest.SourceTypeWorkout))
	})
	bus.Subscribe(event.WeightLogged, func(ctx context.Context, e event.Event) {
		_ = qstSvc.HandleSourceEvent(ctx, e.UserID, string(quest.SourceTypeWeight))
	})

	portH := portfolioHandler.New()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(recover.New())
	app.Use(middleware.RequestLogger(logger))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000,https://kinkando-dev.pages.dev,https://kinkando.dev,https://cronjob.kinkandojester.workers.dev",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api/v1")

	userGroup := api.Group("/users", authMW.Require())
	usrH.Register(userGroup)

	financeGroup := api.Group("/finance", authMW.Require())
	finH.Register(financeGroup)

	kanbanGroup := api.Group("/kanban", authMW.Require())
	kanH.Register(kanbanGroup)

	healthGroup := api.Group("/health", authMW.Require())
	heaH.Register(healthGroup)

	workoutGroup := api.Group("/workout", authMW.Require())
	wkH.Register(workoutGroup)

	medicineGroup := api.Group("/medicines", authMW.Require())
	medH.Register(medicineGroup)

	questGroup := api.Group("/quest", authMW.Require())
	qstH.Register(questGroup)

	portfolioGroup := api.Group("/portfolio")
	portH.Register(portfolioGroup)

	// Resolve the single app user for LINE webhook (and MCP if enabled).
	// MCP_USER_FIREBASE_UID is required; the server refuses to start if unresolvable.
	appUserUUID, err := usrRepo.GetIDByFirebaseUID(ctx, cfg.MCPUserFirebaseUID)
	if err != nil {
		logger.Fatal("could not resolve MCP_USER_FIREBASE_UID",
			zap.String("uid", cfg.MCPUserFirebaseUID), zap.Error(err))
	}
	if appUserUUID == (uuid.UUID{}) {
		logger.Fatal("user not found in users table; sign in via the web app first",
			zap.String("uid", cfg.MCPUserFirebaseUID))
	}

	// MCP server — used by both the in-process Gemini path and (optionally) the
	// HTTP /mcp endpoint. Created unconditionally so Gemini always has a session.
	mcpSrv := mcpserver.New(mcpserver.Deps{
		FinSvc: finSvc, KanRepo: kanRepo, WkSvc: wkSvc, HeaSvc: heaSvc, MedSvc: medSvc,
		UserUUID: appUserUUID, FirebaseUID: cfg.MCPUserFirebaseUID,
		Logger: logger,
	})

	// Wire an in-process MCP client so Gemini can call tools without a network hop.
	serverT, clientT := mcp.NewInMemoryTransports()
	mcpServerSession, err := mcpSrv.Connect(context.Background(), serverT, nil)
	if err != nil {
		logger.Fatal("mcp server session", zap.Error(err))
	}
	defer mcpServerSession.Close() //nolint:errcheck
	mcpCli := mcp.NewClient(&mcp.Implementation{Name: "kinkando-line", Version: "0.1.0"}, nil)
	mcpClientSession, err := mcpCli.Connect(context.Background(), clientT, nil)
	if err != nil {
		logger.Fatal("mcp client session", zap.Error(err))
	}
	defer mcpClientSession.Close() //nolint:errcheck

	// LINE webhook — no auth middleware; self-authenticated via X-Line-Signature.
	lineClient := line.NewClient(cfg.LineChannelAccessToken)

	geminiClient, err := gemini.New(context.Background(), gemini.Deps{
		APIKey:   cfg.GeminiAPIKey,
		Model:    cfg.GeminiModel,
		TTSModel: cfg.GeminiTTSModel,
		MCP:      mcpClientSession,
	})
	if err != nil {
		logger.Fatal("gemini init", zap.Error(err))
	}
	logger.Info("Gemini AI routing enabled for LINE webhook", zap.String("model", cfg.GeminiModel))

	lineH := lineHandler.New(lineHandler.Deps{
		ChannelID:     cfg.LineChannelID,
		ChannelSecret: cfg.LineChannelSecret,
		Client:        lineClient,
		Gemini:        geminiClient,
		Linker:        usrRepo,
		Logger:        logger,
	})
	lineGroup := api.Group("/line")
	lineH.Register(lineGroup)
	logger.Info("LINE webhook enabled at /api/v1/line/webhook")

	// AI chat — authenticated web-app endpoint; reuses the same Gemini+MCP pipeline.
	aiChatH := aichatHandler.New(aichatHandler.Deps{Gemini: geminiClient, Logger: logger})
	aiChatGroup := api.Group("/ai-chat", authMW.Require())
	aiChatH.Register(aiChatGroup)
	logger.Info("AI chat enabled at /api/v1/ai-chat")

	if cfg.MCPAuthToken != "" {
		h := mcp.NewStreamableHTTPHandler(
			func(*http.Request) *mcp.Server { return mcpSrv },
			&mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true},
		)
		app.All("/mcp", mcpBearerAuth(cfg.MCPAuthToken), adaptor.HTTPHandler(h))
		logger.Info("MCP enabled at /mcp", zap.String("user", cfg.MCPUserFirebaseUID))
	} else {
		logger.Info("MCP disabled (set MCP_AUTH_TOKEN to enable)")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		logger.Info("server starting", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("shutting down server")
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if err := app.ShutdownWithContext(shutCtx); err != nil {
		logger.Error("shutdown error", zap.Error(err))
	}
}

func mcpBearerAuth(token string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		got := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
		if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid MCP token")
		}
		return c.Next()
	}
}
