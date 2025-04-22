package main

import (
    "fmt"
    "log"
    "context"
)

var ctx = context.Background()

func CheckHosts(cfg *Config) {
    fmt.Println("Проверка роли для FQDN:", cfg.ClusterFQDN)
    conn, err := ConnectToHost(cfg, cfg.ClusterFQDN)
    if err != nil {
        log.Printf("[FQDN] Ошибка подключения: %v\n", err)
        return
    }
    defer conn.Close(ctx)

    role, err := GetRole(conn)
    if err != nil {
        log.Printf("[FQDN] Ошибка определения роли: %v\n", err)
    } else {
        fmt.Printf("[FQDN] Роль: %s\n", role)
    }

    for _, host := range cfg.Hosts {
        conn, err := ConnectToHost(cfg, host)
        if err != nil {
            log.Printf("[ХОСТ %s] Ошибка подключения: %v\n", host, err)
            continue
        }
        defer conn.Close(ctx)

        role, err := GetRole(conn)
        if err != nil {
            log.Printf("[ХОСТ %s] Ошибка определения роли: %v\n", host, err)
        } else {
            fmt.Printf("[ХОСТ %s] Роль: %s\n", host, role)
        }
    }
}
