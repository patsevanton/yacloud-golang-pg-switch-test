package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

func CheckHosts(cfg *Config) {
	fmt.Println("проверка через hosts:")
	for _, host := range cfg.Hosts {
		hostIPs, err := net.LookupIP(host)
		if err != nil {
			log.Printf("[ХОСТ %s] Ошибка получения IP: %v\n", host, err)
			continue
		}

		pool, dsn, err := ConnectToHost(cfg, host)
		if err != nil {
			if strings.Contains(err.Error(), "read only connection") {
				// Не выводим ничего, если соединение read-only
				continue
			}
			log.Printf("[ХОСТ %s] Ошибка подключения: %v\n", host, err)
			continue
		}
		defer pool.Close()

		role, err := GetRole(pool)
		if err != nil {
			log.Printf("[ХОСТ %s] Ошибка определения роли: %v\n", host, err)
			continue
		}

		var ips []string
		for _, ip := range hostIPs {
			ips = append(ips, ip.String())
		}
		fmt.Printf("role %s через hosts: %s(%s)\n", role, host, strings.Join(ips, ","))
		fmt.Printf("DSN: %s\n", hidePasswordInDSN(dsn))
	}
}
