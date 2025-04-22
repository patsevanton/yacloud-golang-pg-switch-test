package main

import (
    "log"
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

    CheckHosts(cfg)
}
