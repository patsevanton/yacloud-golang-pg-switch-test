package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertCheckRecord(pool *pgxpool.Pool, host string) (bool, error) {
	message := fmt.Sprintf("Проверка подключения к %s в %s", host, time.Now().Format("2006-01-02 15:04:05"))

	_, err := pool.Exec(context.Background(), `
		INSERT INTO health_check (message) VALUES ($1)
	`, message)

	if err != nil {
		return false, err
	}

	return true, nil
}
