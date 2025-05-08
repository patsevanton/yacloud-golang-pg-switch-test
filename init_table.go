package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDatabase(cfg *Config) error {
	fmt.Println("Инициализация базы данных...")

	var pool *pgxpool.Pool
	var err error

	pool, _, err = ConnectToHost(cfg, cfg.ClusterFQDN)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к базе данных для инициализации: %v", err)
	}
	defer pool.Close()

	role, err := GetRole(pool)
	if err != nil {
		return fmt.Errorf("ошибка определения роли: %v", err)
	}

	if role != "master" {
		return fmt.Errorf("для инициализации таблицы требуется подключение к мастеру")
	}

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