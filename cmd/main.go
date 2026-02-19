package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/banking-superapp/upi-service/config"
	"github.com/banking-superapp/upi-service/handler"
	"github.com/banking-superapp/upi-service/repository"
	"github.com/banking-superapp/upi-service/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	cfg := config.Load()

	mongoClient, err := repository.NewMongoClient(cfg.MongoAtlasURI)
	if err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	db := mongoClient.Database("banking_upi")
	if err := repository.CreateIndexes(db); err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	vpaRepo := repository.NewVPARepo(db)
	txnRepo := repository.NewTxnRepo(db)
	mandateRepo := repository.NewMandateRepo(db)
	collectRepo := repository.NewCollectRepo(db)

	upiSvc := service.NewUPIService(vpaRepo, txnRepo, mandateRepo, collectRepo)
	upiHandler := handler.NewUPIHandler(upiSvc)

	app := fiber.New(fiber.Config{
		AppName:      cfg.ServiceName,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": cfg.ServiceName})
	})

	v1 := app.Group("/v1")
	upi := v1.Group("/upi")
	upi.Post("/vpa/create", upiHandler.CreateVPA)
	upi.Get("/vpa", upiHandler.GetVPAs)
	upi.Post("/validate", upiHandler.ValidateVPA)
	upi.Post("/pay", upiHandler.Pay)
	upi.Post("/collect", upiHandler.Collect)
	upi.Get("/transactions", upiHandler.GetTransactions)
	upi.Post("/mandate/create", upiHandler.CreateMandate)
	upi.Get("/mandate", upiHandler.GetMandates)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting %s on port %s", cfg.ServiceName, cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = app.ShutdownWithContext(ctx)
}
