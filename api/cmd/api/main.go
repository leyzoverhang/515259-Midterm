package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"wongnok/internal/auth"
	"wongnok/internal/config"
	"wongnok/internal/middleware"
	"wongnok/internal/platform/cache"
	"wongnok/internal/platform/database"
	"wongnok/internal/user"

	_ "wongnok/docs"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title			Wongnok API
//	@version		1.0
//	@description	API สำหรับจัดการกับระบบสูตรอาหาร
//	@host			localhost:8080
//	@BasePath		/api/v1
//	@schemas		http https

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				พิมพ์ "Bearer" ตามด้วย space แล้วตามด้วย JWT token เช่น "Bearer eyJhbGci..."
func main() {
	if err := run(); err != nil {
		slog.Error("service stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Default logger
	slog.SetDefault(newLogger(os.Stdout, "wongnok", config.Logging{Level: "DEBUG", Format: "text"}))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration:\n%s", config.Humanize(err))
	}

	// Setup logger
	slog.SetDefault(newLogger(os.Stdout, cfg.App.Name, cfg.Logging))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Database connection
	db, sqldb, err := database.Open(ctx, cfg.Database.PostgresDSN)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer sqldb.Close()

	// Redis connection
	rdb, err := cache.Open(ctx, cfg.Redis)
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer rdb.Close()

	// Provider
	oidcProvider, err := oidc.NewProvider(ctx, cfg.Keycloak.RealmURL())
	if err != nil {
		return fmt.Errorf("discover keycloak provider: %w", err)
	}

	oidcVerifier := oidcProvider.Verifier(&oidc.Config{ClientID: cfg.Keycloak.ClientID})

	// Dependency injection
	authRepo := auth.NewRepository(rdb)
	authService := auth.NewService(authRepo, cfg.Keycloak, oidcProvider)
	authHandler := auth.NewHandler(authService)

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	// Router
	router := gin.Default()

	// Use global middleware
	router.Use(cors.Default())

	// Group version
	v1 := router.Group("/api/v1")

	// Auth resource
	authGroup := v1.Group("/auth")
	authGroup.GET("/login", authHandler.Login)
	authGroup.GET("/callback", authHandler.Cabllback)
	authGroup.POST("/exchange", authHandler.Exchange)
	authGroup.POST("/logout", authHandler.Logout)

	// User resource
	userGroup := v1.Group("/users")

	// User JWT middleware
	userGroup.Use(middleware.JWT(oidcVerifier))

	// Register path
	// curl -X GET http://localhost:8080/api/v1/users/{id}
	userGroup.GET("/:id", userHandler.GetUser)

	// curl -X POST http://localhost:8080/api/v1/users -H "Content-Type: application/json" -d '{"email":"taro@devpool.pea"}'
	userGroup.POST("", userHandler.CreateUser)

	// Register swagger
	router.GET("swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Server
	serv := &http.Server{
		Addr:    cfg.App.Addr(),
		Handler: router,
	}

	go serv.ListenAndServe()
	log.Printf("server started at %s\n", cfg.App.Addr())

	// Graceful shutdown
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	return serv.Shutdown(shutdownCtx)
}

func newLogger(writer io.Writer, name string, log config.Logging) *slog.Logger {
	opts := &slog.HandlerOptions{Level: log.SlogLevel()}

	var handler slog.Handler = slog.NewJSONHandler(writer, opts)
	if log.Format == "text" {
		handler = slog.NewTextHandler(writer, opts)
	}

	return slog.New(handler).With("service", name)
}
