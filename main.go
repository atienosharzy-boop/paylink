package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"paylink/internal/config"
	"paylink/internal/handlers"
	"paylink/internal/models"
	"paylink/internal/mpesa"
	"paylink/internal/sms"
	"paylink/internal/store"
	"syscall"
	"time"
)

//go:embed web/templates/* web/static/*
var embeddedFiles embed.FS

func main() {
	log.Println("Starting PayLink for Schools & Academies...")

	cfg := config.Load()

	// Initialize persistent storage (data/paylink_db.json)
	dataStore, err := store.NewStore("data/paylink_db.json", cfg.School)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	// Prepare template funcs
	tmplFuncs := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"formatMoney": func(v int64) string {
			return models.FormatMoney(v)
		},
	}

	// Parse embedded HTML templates
	tmpl, err := template.New("").Funcs(tmplFuncs).ParseFS(embeddedFiles, "web/templates/*.html")
	if err != nil {
		log.Fatalf("Template parsing failed: %v", err)
	}

	// Static filesystem
	staticSubFS, err := fs.Sub(embeddedFiles, "web/static")
	if err != nil {
		log.Fatalf("Static FS creation failed: %v", err)
	}

	// Initialize services
	mpesaClient := mpesa.NewClient(cfg)
	smsService := sms.NewService(cfg)

	// Initialize handlers
	dashboardHandler := handlers.NewDashboardHandler(cfg, dataStore, smsService, tmpl)
	paymentHandler := handlers.NewPaymentHandler(cfg, dataStore, mpesaClient, smsService, tmpl)

	// Initialize Router
	router := handlers.NewRouter(dashboardHandler, paymentHandler, staticSubFS)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background goroutine
	go func() {
		fmt.Println("================================================================")
		fmt.Printf(" PayLink School Fee Tracker is live at http://localhost:%s\n", cfg.Port)
		fmt.Println("================================================================")
		fmt.Printf(" School Name:    %s\n", cfg.School.Name)
		fmt.Printf(" Current Term:   %s\n", cfg.School.CurrentTerm)
		fmt.Printf(" School Paybill: %s\n", cfg.School.Paybill)
		fmt.Printf(" Mode:           %s\n", func() string {
			if cfg.DemoMode {
				return "⚡ Demo / Simulation (Automatic M-Pesa push approval in 3s)"
			}
			return "Production (Live Daraja & Africa's Talking)"
		}())
		fmt.Println("----------------------------------------------------------------")
		fmt.Printf(" 🏠 Public Landing Page: http://localhost:%s/\n", cfg.Port)
		fmt.Printf(" 📊 Bursar Dashboard:    http://localhost:%s/admin\n", cfg.Port)
		fmt.Printf(" 📱 Parent Fee Link:     http://localhost:%s/pay/amani-kimani-adm-2024-012\n", cfg.Port)
		fmt.Println("================================================================")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("PayLink stopped cleanly.")
}
