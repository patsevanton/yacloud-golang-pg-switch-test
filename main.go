package main

import (
    "fmt"
	"log"
	"time"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Бесконечный цикл с проверками каждые 5 секунд
	for {
		CheckHosts(cfg)
		time.Sleep(5 * time.Second)
		fmt.Println() // Пустая строка между выводами
	}
}
