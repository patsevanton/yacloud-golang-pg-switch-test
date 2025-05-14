package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertCheckRecord(pool *pgxpool.Pool, host string) (bool, error) {
	message := fmt.Sprintf("Проверка подключения к %s в %s", host, time.Now().Format("2006-01-02 15:04:05"))
	_, err := pool.Exec(context.Background(), `INSERT INTO health_check (message) VALUES ($1)`, message)
	return err == nil, err
}

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