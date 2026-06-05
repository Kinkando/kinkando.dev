package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/kinkando/personal-dashboard/config"
	aichatHandler "github.com/kinkando/personal-dashboard/internal/aichat/handler"
	"github.com/kinkando/personal-dashboard/internal/auth"
	"github.com/kinkando/personal-dashboard/internal/fcm"
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
	cronHandler "github.com/kinkando/personal-dashboard/internal/cron/handler"
	healthReminder "github.com/kinkando/personal-dashboard/internal/health/reminder"
	medicineHandler "github.com/kinkando/personal-dashboard/internal/medicine/handler"
	medicineRepo "github.com/kinkando/personal-dashboard/internal/medicine/repository"
	medReminder "github.com/kinkando/personal-dashboard/internal/medicine/reminder"
	medicineSvc "github.com/kinkando/personal-dashboard/internal/medicine/service"
	questReminder  "github.com/kinkando/personal-dashboard/internal/quest/reminder"
	questSnapshot  "github.com/kinkando/personal-dashboard/internal/quest/snapshot"
	"github.com/kinkando/personal-dashboard/internal/reminderlog"
	"github.com/kinkando/personal-dashboard/internal/notification"
	notificationHandler "github.com/kinkando/personal-dashboard/internal/notification/handler"
	notificationRepo "github.com/kinkando/personal-dashboard/internal/notification/repository"
	notificationSvc "github.com/kinkando/personal-dashboard/internal/notification/service"
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

	// LINE client — constructed early so it can be shared with the notification module.
	lineClient := line.NewClient(cfg.LineChannelAccessToken)

	// Notification module (LINE push, Discord webhook, FCM web push).
	fcmClient, err := fcm.NewClient(ctx, cfg.FirebaseCredentials)
	if err != nil {
		logger.Fatal("fcm init", zap.Error(err))
	}
	discordClient := notification.NewDiscordClient()
	notiRepo := notificationRepo.New(pgDB.SQL())
	notiSvc := notificationSvc.New(notiRepo, lineClient, discordClient, fcmClient, usrRepo, logger)
	notiH := notificationHandler.New(notiSvc, usrRepo)

	remLogRepo := reminderlog.New(pgDB.SQL())

	medRemSvc := medReminder.New(medRepo, notiSvc, logger)
	qstRemSvc := questReminder.New(qstRepo, remLogRepo, notiSvc, logger)
	qstSnapSvc := questSnapshot.New(qstRepo, logger)
	heaRemSvc := healthReminder.New(heaRepo, remLogRepo, notiSvc, logger)

	// Recommended Cloudflare Worker cron schedules (UTC; Bangkok = UTC+7):
	//
	//   medicine-reminders       */30 * * * *      every 30 min all day
	//                                               dose window = 30 min; supply digest self-gates ≥ 09:00 BKK (02:00 UTC)
	//
	//   quest-reminders          */30 * * * *      every 30 min all day
	//                                               daily quest nudge self-gates ≥ 20:00 BKK (13:00 UTC)
	//                                               weekly quest nudge self-gates Sunday ≥ 18:00 BKK (11:00 UTC)
	//
	//   weight-nudge             0,30 1-3 * * *    08:00–10:30 BKK (01:00–03:30 UTC)
	//                                               self-gates ≥ 08:00 BKK; once per user per day via dedup
	//
	//   quest-period-snapshot    0,30 17 * * *     00:00 & 00:30 BKK (17:00 & 17:30 UTC)
	//                                               daily snapshot every day (records yesterday);
	//                                               weekly snapshot on Mondays only (records the just-ended week)
	//                                               idempotent via quest_period_results unique constraint
	//
	// All endpoints are idempotent — repeat runs within the same period are safe.
	cronH := cronHandler.New(map[string]cronHandler.RunFunc{
		"medicine-reminders":    func(ctx context.Context) (any, error) { return medRemSvc.Run(ctx) },
		"quest-reminders":       func(ctx context.Context) (any, error) { return qstRemSvc.Run(ctx) },
		"quest-period-snapshot": func(ctx context.Context) (any, error) { return qstSnapSvc.Run(ctx) },
		"weight-nudge":          func(ctx context.Context) (any, error) { return heaRemSvc.Run(ctx) },
	})

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
	bus.Subscribe(event.SleepLogged, func(ctx context.Context, e event.Event) {
		_ = qstSvc.HandleSourceEvent(ctx, e.UserID, string(quest.SourceTypeSleep))
	})

	// Subscribe notification service to domain events for fan-out delivery.
	bus.Subscribe(event.WeightLogged, func(ctx context.Context, e event.Event) {
		notiSvc.Notify(ctx, e.UserID, notification.Message{Title: "Weight logged", Body: "Your weight log has been recorded."})
	})
	bus.Subscribe(event.SleepLogged, func(ctx context.Context, e event.Event) {
		notiSvc.Notify(ctx, e.UserID, notification.Message{Title: "Sleep logged", Body: "Your sleep log has been recorded."})
	})
	bus.Subscribe(event.MedicineTaken, func(ctx context.Context, e event.Event) {
		notiSvc.Notify(ctx, e.UserID, notification.Message{Title: "Medicine taken", Body: "Your medicine intake has been recorded."})
	})
	bus.Subscribe(event.SupplementTaken, func(ctx context.Context, e event.Event) {
		notiSvc.Notify(ctx, e.UserID, notification.Message{Title: "Supplement taken", Body: "Your supplement intake has been recorded."})
	})
	bus.Subscribe(event.WorkoutSessionFinished, func(ctx context.Context, e event.Event) {
		notiSvc.Notify(ctx, e.UserID, notification.Message{Title: "Workout complete", Body: "Great job! Your workout session has been recorded."})
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
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Cron-Secret",
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

	notificationGroup := api.Group("/notifications", authMW.Require())
	notiH.Register(notificationGroup)

	portfolioGroup := api.Group("/portfolio")
	portH.Register(portfolioGroup)

	// Cron endpoints — authenticated by shared secret, not Firebase.
	// Called by the Cloudflare cron worker at cronjob.kinkandojester.workers.dev.
	cronGroup := api.Group("/cron", middleware.CronAuth(cfg.CronSecret))
	cronH.Register(cronGroup)

	// Single shared MCP server — a receiving middleware resolves the caller per
	// request from either the HTTP header (X-MCP-User, set by mcpFirebaseAuth)
	// or the in-process call's _meta (set by the Gemini client).
	mcpSrv := mcpserver.New(mcpserver.Deps{
		FinSvc: finSvc, KanRepo: kanRepo, WkSvc: wkSvc, HeaSvc: heaSvc, MedSvc: medSvc, QstSvc: qstSvc,
		Resolver: usrRepo, Logger: logger,
	})

	// Wire an in-process MCP client so Gemini can call tools without a network hop.
	serverT, clientT := mcp.NewInMemoryTransports()
	mcpServerSession, err := mcpSrv.Connect(context.Background(), serverT, nil)
	if err != nil {
		logger.Fatal("mcp server session", zap.Error(err))
	}
	defer mcpServerSession.Close() //nolint:errcheck
	mcpCli := mcp.NewClient(&mcp.Implementation{Name: "kinkando-in-process", Version: "0.1.0"}, nil)
	mcpClientSession, err := mcpCli.Connect(context.Background(), clientT, nil)
	if err != nil {
		logger.Fatal("mcp client session", zap.Error(err))
	}
	defer mcpClientSession.Close() //nolint:errcheck

	// LINE webhook — no auth middleware; self-authenticated via X-Line-Signature.
	// lineClient is constructed above (shared with the notification module).

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
		Users:         usrRepo,
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

	// MCP endpoint — always mounted; authenticated per-request via Firebase ID token.
	// The caller must send: Authorization: Bearer <firebase-id-token>.
	// mcpFirebaseAuth verifies the token, sets X-MCP-User, then the MCP server's
	// receiving middleware resolves the user from that header per tool call.
	mcpH := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return mcpSrv },
		&mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true},
	)
	app.All("/mcp", mcpFirebaseAuth(authMW), adaptor.HTTPHandler(mcpH))
	logger.Info("MCP enabled at /mcp (Firebase ID token required)")

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

// mcpFirebaseAuth verifies the Firebase ID token supplied as "Authorization: Bearer <token>"
// and stashes the resolved UID in the X-MCP-User request header so the downstream
// net/http handler (via adaptor) can access it after the Fiber context is gone.
func mcpFirebaseAuth(mw *auth.Middleware) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if len(header) < 8 || header[:7] != "Bearer " {
			return fiber.NewError(fiber.StatusUnauthorized, "missing firebase bearer token")
		}
		idToken := header[7:]
		uid, _, err := mw.Verify(c.Context(), idToken)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired firebase token")
		}
		// Stash the UID so the net/http handler can read it after the adaptor.
		c.Request().Header.Set("X-MCP-User", uid)
		return c.Next()
	}
}
