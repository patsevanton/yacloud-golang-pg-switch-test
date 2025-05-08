package main

import (
	"os"
	"strings"
	"fmt"
)

type Config struct {
	PGUser      string
	PGPassword  string
	PGDatabase  string
	ClusterFQDN string
}

func LoadConfig() (*Config, error) {
	if os.Getenv("CLUSTER_FQDN") == "" {
		return nil, fmt.Errorf("переменная CLUSTER_FQDN не задана")
	}

	return &Config{
		PGUser:      os.Getenv("PG_USER"),
		PGPassword:  os.Getenv("PG_PASSWORD"),
		PGDatabase:  os.Getenv("PG_DB"),
		ClusterFQDN: strings.TrimSpace(os.Getenv("CLUSTER_FQDN")),
	}, nil
}
