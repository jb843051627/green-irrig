package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"green-irrig/internal/handler"
	"green-irrig/internal/store"
	"green-irrig/internal/service"
)

func main() {
	db, err := store.OpenDB()
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	svc := service.NewIrrigationService(db)
	h := handler.NewIrrigationHandler(svc)

	r := gin.Default()
	h.RegisterRoutes(r)

	log.Println("green-irrig starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
