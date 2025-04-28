package main

import (
	"github.com/jackc/pgx/v5"
)

func GetRole(conn *pgx.Conn) (string, error) {
	var isInRecovery bool
	err := conn.QueryRow(ctx, "SELECT pg_is_in_recovery()").Scan(&isInRecovery)
	if err != nil {
		return "", err
	}
	if isInRecovery {
		return "replica", nil
	}
	return "master", nil
}
