package main

import (
	"fmt"
	"net"
)

func CheckHosts(cfg *Config) {
	fmt.Println("проверка через hosts:")
	for _, host := range cfg.Hosts {
		hostIPs, err := net.LookupIP(host)
		if err != nil {
			continue
		}

		pool, _, err := ConnectToHost(cfg, host)
		if err != nil {
			continue
		}
		defer pool.Close()

		role, err := GetRole(pool)
		if err != nil {
			continue
		}

		if len(hostIPs) == 0 {
			continue
		}

		fmt.Printf("%s: %s(%s)\n", role, host, hostIPs[0].String())

		// Добавляем проверку вставки
		success, err := InsertCheckRecord(pool, host)
		if err != nil {
			fmt.Printf("Ошибка вставки для %s: %v\n", host, err)
		} else if success {
			fmt.Printf("insert successful для %s\n", host)
		}
	}
}
