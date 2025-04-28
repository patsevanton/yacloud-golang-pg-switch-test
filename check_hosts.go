package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

func CheckHosts(cfg *Config) {
	for _, host := range cfg.Hosts {
		hostIPs, err := net.LookupIP(host)
		if err != nil {
			log.Printf("[ХОСТ %s] Ошибка получения IP: %v\n", host, err)
			continue
		}

		conn, dsn, err := ConnectToHost(cfg, host)
		if err != nil {
			if strings.Contains(err.Error(), "read only connection") {
				continue
			}
			log.Printf("[ХОСТ %s] Ошибка подключения: %v\n", host, err)
			continue
		}
		defer conn.Close(ctx)

		role, err := GetRole(conn)
		if err != nil {
			log.Printf("[ХОСТ %s] Ошибка определения роли: %v\n", host, err)
		} else {
			var ips []string
			for _, ip := range hostIPs {
				ips = append(ips, ip.String())
			}
			fmt.Printf("%s через libpq: %s(%s)\n", role, host, strings.Join(ips, ","))
			fmt.Printf("DSN: %s\n", hidePasswordInDSN(dsn))
		}
	}
}
