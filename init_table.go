package main

import (
	"context"
	"fmt"
// 	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InitDatabase создает таблицу для проверки, если она не существует
func InitDatabase(cfg *Config) error {
	fmt.Println("Инициализация базы данных...")

	var pool *pgxpool.Pool
	var err error

	// Сначала попробуем подключиться через ClusterFQDN
	pool, _, err = ConnectToHost(cfg, cfg.ClusterFQDN)
	if err != nil {
		// Если не удалось подключиться через ClusterFQDN, попробуем через хосты
		for _, host := range cfg.Hosts {
			pool, _, err = ConnectToHost(cfg, host)
			if err == nil {
				break
			}
		}
	}

	if err != nil {
		return fmt.Errorf("не удалось подключиться к базе данных для инициализации: %v", err)
	}
	defer pool.Close()

	// Получаем роль, чтобы убедиться, что мы подключены к мастеру
	role, err := GetRole(pool)
	if err != nil {
		return fmt.Errorf("ошибка определения роли: %v", err)
	}

	if role != "master" {
		return fmt.Errorf("для инициализации таблицы требуется подключение к мастеру")
	}

	// Создаем таблицу, если она не существует
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS health_check (
			id SERIAL PRIMARY KEY,
			check_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			message TEXT
		)
	`)

	if err != nil {
		return fmt.Errorf("ошибка создания таблицы: %v", err)
	}

	fmt.Println("Таблица health_check успешно инициализирована")
	return nil
}
