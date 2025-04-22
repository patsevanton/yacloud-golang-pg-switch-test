package main

import (
    "fmt"
    "github.com/jackc/pgx/v5"
)

func ConnectToHost(cfg *Config, host string) (*pgx.Conn, error) {
    dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=require",
        cfg.PGUser, cfg.PGPassword, host, cfg.PGDB)
    return pgx.Connect(ctx, dsn)
}

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
