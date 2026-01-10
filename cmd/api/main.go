package main

import (
	"TugasAkhir/config"
	"TugasAkhir/routes"
	"TugasAkhir/utils/fcm"
	"TugasAkhir/utils/storage"
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	config.LoadEnv()

	if err := config.Validate(); err != nil {
		log.Fatalf("configuration validation failed: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config.ConnectDB()
	storage.InitS3Client()
	app := fiber.New()

	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "capacitor://localhost,ionic://localhost,http://localhost,https://localhost",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))

	routes.SetupRoutes(app, config.DB)

	go func() {
		log.Println("🚀 API running on :8080")
		go fcm.StartNotifierConsumer(ctx)
		if err := app.Listen(":8080"); err != nil && !errors.Is(err, net.ErrClosed) {
			log.Fatalf("fiber server error: %v", err)
		}
	}()

	<-ctx.Done()
	stop()

	log.Println("⌛ shutting down server ...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("error during server shutdown: %v", err)
	}
	log.Println("✅ server gracefully stopped")
}
