package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetRole(pool *pgxpool.Pool) (string, error) {
	var isInRecovery bool
	err := pool.QueryRow(context.Background(), "SELECT pg_is_in_recovery()").Scan(&isInRecovery)
	if err != nil {
		return "", err
	}
	if isInRecovery {
		return "replica", nil
	}
	return "master", nil
}
