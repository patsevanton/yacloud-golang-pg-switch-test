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
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализируем базу данных в начале работы
	err = InitDatabase(cfg)
	if err != nil {
		log.Printf("Предупреждение: %v", err)
		// Продолжаем работу даже если не удалось создать таблицу
	}

	// Бесконечный цикл с проверками каждые 5 секунд
	for {
		fmt.Printf("\n=== Проверка %s ===\n", time.Now().Format("2006-01-02 15:04:05"))
		CheckHosts(cfg)
		fmt.Println()
		CheckClusterFQDN(cfg)
		time.Sleep(5 * time.Second)
		fmt.Println()
	}
}
