package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Глобальный контекст
var ctx = context.Background()

// Config содержит настройки подключения к PostgreSQL
type Config struct {
	PGUser               string
	PGPassword           string
	PGDatabase           string
	ClusterFQDN          string
	PGSSLMode            string
	PGTargetSessionAttrs string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() (*Config, error) {
	requiredVars := map[string]string{
		"CLUSTER_FQDN":            "переменная CLUSTER_FQDN не задана",
		"PG_SSLMODE":              "переменная PG_SSLMODE не задана",
		"PG_TARGET_SESSION_ATTRS": "переменная PG_TARGET_SESSION_ATTRS не задана",
	}

	for env, msg := range requiredVars {
		if os.Getenv(env) == "" {
			return nil, fmt.Errorf(msg)
		}
	}

	return &Config{
		PGUser:               os.Getenv("PG_USER"),
		PGPassword:           os.Getenv("PG_PASSWORD"),
		PGDatabase:           os.Getenv("PG_DB"),
		ClusterFQDN:          strings.TrimSpace(os.Getenv("CLUSTER_FQDN")),
		PGSSLMode:            os.Getenv("PG_SSLMODE"),
		PGTargetSessionAttrs: os.Getenv("PG_TARGET_SESSION_ATTRS"),
	}, nil
}

// ConnectToPostgreSQL устанавливает соединение с базой данных PostgreSQL
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

// InitDatabase инициализирует таблицу health_check в базе данных
func InitDatabase(cfg *Config) error {
	fmt.Println("Инициализация базы данных...")

	pool, _, err := ConnectToPostgreSQL(cfg, cfg.ClusterFQDN)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к базе данных для инициализации: %v", err)
	}
	defer pool.Close()

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

// InsertCheckRecord добавляет запись о проверке в таблицу health_check
func InsertCheckRecord(pool *pgxpool.Pool, host string) (bool, error) {
	message := fmt.Sprintf("Проверка подключения к %s в %s", host, time.Now().Format("2006-01-02 15:04:05"))
	_, err := pool.Exec(context.Background(), `INSERT INTO health_check (message) VALUES ($1)`, message)
	return err == nil, err
}

// CheckClusterFQDN проверяет соединение с кластером и записывает результат
func CheckClusterFQDN(cfg *Config) {
	fqdnIPs, err := net.LookupIP(cfg.ClusterFQDN)
	if err != nil || len(fqdnIPs) == 0 {
		return
	}

	if cname, err := net.LookupCNAME(cfg.ClusterFQDN); err == nil && cname != cfg.ClusterFQDN {
		fmt.Printf("%s cname на хост %s\n", cfg.ClusterFQDN, cname)
	}

	pool, _, err := ConnectToPostgreSQL(cfg, cfg.ClusterFQDN)
	if err != nil {
		return
	}
	defer pool.Close()

	success, err := InsertCheckRecord(pool, cfg.ClusterFQDN)
	if err != nil {
		fmt.Printf("Ошибка вставки для %s: %v\n", cfg.ClusterFQDN, err)
	} else if success {
		fmt.Printf("insert successful для %s\n", cfg.ClusterFQDN)
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	if err := InitDatabase(cfg); err != nil {
		log.Printf("Предупреждение: %v", err)
	}

	for {
		fmt.Printf("\n=== Проверка %s ===\n", time.Now().Format("2006-01-02 15:04:05"))
		CheckClusterFQDN(cfg)
		time.Sleep(500 * time.Millisecond)
		fmt.Println()
	}
}