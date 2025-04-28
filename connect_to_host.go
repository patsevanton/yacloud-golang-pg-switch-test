package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
)

func ConnectToHost(cfg *Config, host string) (*pgx.Conn, string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, "", fmt.Errorf("не удалось получить путь к исполняемому файлу: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	certPath := filepath.Join(exeDir, "yandexcloud.crt")

	caCert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, "", fmt.Errorf("unable to read CA cert: %v (path: %s)", err, certPath)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, "", fmt.Errorf("failed to add CA cert to pool")
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:6432/%s?sslmode=verify-full&target_session_attrs=read-write",
		cfg.PGUser, cfg.PGPassword, host, cfg.PGDB)

	connConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, dsn, fmt.Errorf("unable to parse config: %v", err)
	}

	connConfig.TLSConfig = &tls.Config{
		RootCAs:    caCertPool,
		ServerName: host,
	}

	conn, err := pgx.ConnectConfig(ctx, connConfig)
	return conn, dsn, err
}
