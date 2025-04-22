package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func ConnectToHost(cfg *Config, host string) (*pgx.Conn, error) {
	// Загрузка CA сертификата
	caCert, err := os.ReadFile("/home/user/.postgresql/root.crt") // Укажите правильный путь к CA cert
	if err != nil {
		return nil, fmt.Errorf("unable to read CA cert: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to add CA cert to pool")
	}

	// Формирование DSN
	dsn := fmt.Sprintf("postgres://%s:%s@%s:6432/%s?sslmode=verify-full&target_session_attrs=read-write",
		cfg.PGUser, cfg.PGPassword, host, cfg.PGDB)

	// Парсинг конфигурации
	connConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config: %v", err)
	}

	// Настройка TLS
	connConfig.TLSConfig = &tls.Config{
		RootCAs:    caCertPool,
		ServerName: host, // Важно для проверки сертификата
	}

	return pgx.ConnectConfig(ctx, connConfig)
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
