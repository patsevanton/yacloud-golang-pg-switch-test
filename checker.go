package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
)

var ctx = context.Background()

func CheckHosts(cfg *Config) {
	// Получаем IP для FQDN
	fqdnIPs, err := net.LookupIP(cfg.ClusterFQDN)
	if err != nil {
		log.Printf("[FQDN] Ошибка получения IP: %v\n", err)
	} else {
		// Получаем CNAME для FQDN
		cnames, err := net.LookupCNAME(cfg.ClusterFQDN)
		if err != nil {
			log.Printf("[FQDN] Ошибка получения CNAME: %v\n", err)
		}

		conn, err := ConnectToHost(cfg, cfg.ClusterFQDN)
		if err != nil {
			log.Printf("[FQDN] Ошибка подключения: %v\n", err)
		} else {
			defer conn.Close(ctx)
			role, err := GetRole(conn)
			if err != nil {
				log.Printf("[FQDN] Ошибка определения роли: %v\n", err)
			} else {
				// Форматируем вывод для FQDN (без IP в первой строке)
				fmt.Printf("%s cname на хост %s\n",
					cfg.ClusterFQDN, cnames)
				// Добавляем использование role в вывод (с IP во второй строке)
				var ips []string
				for _, ip := range fqdnIPs {
					ips = append(ips, ip.String())
				}
				fmt.Printf("%s через libpq: %s(%s)\n",
					role, cfg.ClusterFQDN, strings.Join(ips, ","))
			}
		}
	}

	for _, host := range cfg.Hosts {
		// Получаем IP для хоста
		hostIPs, err := net.LookupIP(host)
		if err != nil {
			log.Printf("[ХОСТ %s] Ошибка получения IP: %v\n", host, err)
			continue
		}

		conn, err := ConnectToHost(cfg, host)
		if err != nil {
			if strings.Contains(err.Error(), "read only connection") {
				// Пропускаем вывод для read-only реплик
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
			// Форматируем вывод для хоста
			var ips []string
			for _, ip := range hostIPs {
				ips = append(ips, ip.String())
			}
			fmt.Printf("%s через libpq: %s(%s)\n",
				role, host, strings.Join(ips, ","))
		}
	}
}
