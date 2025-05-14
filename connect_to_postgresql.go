package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectToPostgreSQL(cfg *Config, host string) (*pgxpool.Pool, string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, "", fmt.Errorf("не удалось получить путь к исполняемому файлу: %v", err)
	}

	certPath := filepath.Join(filepath.Dir(exePath), "yandexcloud.crt")
	caCert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, "", fmt.Errorf("unable to read CA cert: %v (path: %s)", err, certPath)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, "", fmt.Errorf("failed to add CA cert to pool")
	}

	params := url.Values{}
	if cfg.PGSSLMode != "" {
		params.Set("sslmode", cfg.PGSSLMode)
	}
	if cfg.PGTargetSessionAttrs != "" {
		params.Set("target_session_attrs", cfg.PGTargetSessionAttrs)
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:6432/%s?%s&pool_max_conn_lifetime=1h&pool_max_conn_idle_time=30m",
		url.QueryEscape(cfg.PGUser),
		url.QueryEscape(cfg.PGPassword),
		host,
		url.PathEscape(cfg.PGDatabase),
		params.Encode(),
	)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, dsn, fmt.Errorf("unable to parse config: %v", err)
	}

	config.ConnConfig.TLSConfig = &tls.Config{
		RootCAs:    caCertPool,
		ServerName: host,
	}
	config.MinConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, config)

	return pool, dsn, err
}
