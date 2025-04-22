package main

import (
    "fmt"
    "os"
    "strings"
)

type Config struct {
    PGUser      string
    PGPassword  string
    PGDB        string
    ClusterFQDN string
    Hosts       []string
}

func LoadConfig() (*Config, error) {
    hostsRaw := os.Getenv("HOSTS")
    if hostsRaw == "" {
        return nil, fmt.Errorf("переменная HOSTS не задана")
    }

    return &Config{
        PGUser:      os.Getenv("PG_USER"),
        PGPassword:  os.Getenv("PG_PASSWORD"),
        PGDB:        os.Getenv("PG_DB"),
        ClusterFQDN: os.Getenv("CLUSTER_FQDN"),
        Hosts:       strings.Split(hostsRaw, ","),
    }, nil
}
