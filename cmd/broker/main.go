package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/8thgencore/message-broker/internal/app"
	"github.com/8thgencore/message-broker/internal/config"
)

func main() {
	// Парсим флаги командной строки
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	// Загружаем конфигурацию
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаем приложение
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Запускаем приложение
	if err := application.Run(context.Background()); err != nil {
		log.Printf("Failed to run application: %v", err)
		os.Exit(1)
	}
} 