package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

func CheckClusterFQDN(cfg *Config) {
	fqdnIPs, err := net.LookupIP(cfg.ClusterFQDN)
	if err != nil {
		log.Printf("[FQDN] Ошибка получения IP: %v\n", err)
		return
	}

	cnames, err := net.LookupCNAME(cfg.ClusterFQDN)
	if err != nil {
		log.Printf("[FQDN] Ошибка получения CNAME: %v\n", err)
	}

	conn, dsn, err := ConnectToHost(cfg, cfg.ClusterFQDN)
	if err != nil {
		log.Printf("[FQDN] Ошибка подключения: %v\n", err)
		return
	}
	defer conn.Close(ctx)

	role, err := GetRole(conn)
	if err != nil {
		log.Printf("[FQDN] Ошибка определения роли: %v\n", err)
		return
	}

	fmt.Printf("%s cname на хост %s\n", cfg.ClusterFQDN, cnames)
	var ips []string
	for _, ip := range fqdnIPs {
		ips = append(ips, ip.String())
	}
	fmt.Printf("%s через libpq: %s(%s)\n", role, cfg.ClusterFQDN, strings.Join(ips, ","))
	fmt.Printf("DSN: %s\n", hidePasswordInDSN(dsn))
}
