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

    "github.com/gin-gonic/gin"
    _ "github.com/Luckyisrael/solana-wallet-api/docs" 
    "github.com/Luckyisrael/solana-wallet-api/internal/api/handler"
    "github.com/Luckyisrael/solana-wallet-api/internal/api/middleware"
    "github.com/Luckyisrael/solana-wallet-api/internal/config"
    "github.com/Luckyisrael/solana-wallet-api/internal/solana"
    "github.com/Luckyisrael/solana-wallet-api/internal/repo"
    "github.com/Luckyisrael/solana-wallet-api/internal/service/wallet"
    "github.com/Luckyisrael/solana-wallet-api/internal/redis"
    "github.com/Luckyisrael/solana-wallet-api/internal/service/balance"
	"github.com/Luckyisrael/solana-wallet-api/internal/service/transfer"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/joho/godotenv"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

var cfg *config.Config

// @title           Solana Wallet API
// @version         1.0
// @description     Production-ready Solana wallet backend.
// @host            localhost:8080
// @BasePath        /v1
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}

    cfg = config.Load()
    log.Printf("DB URL: %s", cfg.Postgres.URL)

	// DB Pool
    db, err := pgxpool.New(context.Background(), cfg.Postgres.URL)
    if err != nil {
        log.Fatalf("Unable to connect to database: %v", err)
    }
    defer db.Close()

    // Repos
    walletRepo := repo.NewWalletRepo(db)

    // Services
    walletService := wallet.NewService(walletRepo)

    // Initialize Solana + Redis clients
    solanaClient := solana.NewClient(cfg.Solana.RPCEndpoint)
    redisClient := redis.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
    balanceService := balance.NewService(solanaClient, redisClient)
	transferService := transfer.NewService(walletRepo, solanaClient, redisClient, cfg.AES.MasterKey)

	// Gin mode
	gin.SetMode(gin.ReleaseMode)
	if os.Getenv("GIN_MODE") == "debug" {
		gin.SetMode(gin.DebugMode)
	}

    r := gin.New()
    r.Use(middleware.Logger(), middleware.RecoveryWithZap())

    // Health check
    r.GET("/health", handler.Health)
    r.GET("/v1/health", handler.Health)

    // API v1
    v1 := r.Group("/v1")
    {
        v1.POST("/wallets", handler.CreateWallet(walletService))
        v1.GET("/wallets/:address/balance", handler.GetBalance(balanceService))
		v1.POST("/transactions/transfer", handler.Transfer(transferService))
    }
	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt
	waitForShutdown(srv)
}

func waitForShutdown(srv *http.Server) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	log.Println("Server stopped.")
}
