package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
)

var ctx = context.Background()

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	if err := InitDatabase(cfg); err != nil {
		log.Printf("Предупреждение: %v", err)
	}

	for {
		fmt.Printf("\n=== Проверка %s ===\n", time.Now().Format("2006-01-02 15:04:05"))
		CheckClusterFQDN(cfg)
		time.Sleep(500 * time.Millisecond)
		fmt.Println()
	}
}