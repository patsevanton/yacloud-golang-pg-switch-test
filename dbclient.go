package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
)

// ConnectToHost возвращает соединение и использованный DSN
func ConnectToHost(cfg *Config, host string) (*pgx.Conn, string, error) {
	// Получаем путь к директории с исполняемым файлом
	exePath, err := os.Executable()
	if err != nil {
		return nil, "", fmt.Errorf("не удалось получить путь к исполняемому файлу: %v", err)
	}
	exeDir := filepath.Dir(exePath)

	// Формируем путь к сертификату
	certPath := filepath.Join(exeDir, "yandexcloud.crt")

	// Загрузка CA сертификата
	caCert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, "", fmt.Errorf("unable to read CA cert: %v (path: %s)", err, certPath)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, "", fmt.Errorf("failed to add CA cert to pool")
	}

	// Формирование DSN
	dsn := fmt.Sprintf("postgres://%s:%s@%s:6432/%s?sslmode=verify-full&target_session_attrs=read-write",
		cfg.PGUser, cfg.PGPassword, host, cfg.PGDB)

	// Парсинг конфигурации
	connConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, dsn, fmt.Errorf("unable to parse config: %v", err)
	}

	// Настройка TLS
	connConfig.TLSConfig = &tls.Config{
		RootCAs:    caCertPool,
		ServerName: host, // Важно для проверки сертификата
	}

	conn, err := pgx.ConnectConfig(ctx, connConfig)
	return conn, dsn, err
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
