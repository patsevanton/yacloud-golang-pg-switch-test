package main

import (
	"os"
	"strings"
	"fmt"
)

type Config struct {
	PGUser      string
	PGPassword  string
	PGDB        string
	ClusterFQDN string
}

func LoadConfig() (*Config, error) {
	if os.Getenv("CLUSTER_FQDN") == "" {
		return nil, fmt.Errorf("переменная CLUSTER_FQDN не задана")
	}

	return &Config{
		PGUser:      os.Getenv("PG_USER"),
		PGPassword:  os.Getenv("PG_PASSWORD"),
		PGDB:        os.Getenv("PG_DB"),
		ClusterFQDN: strings.TrimSpace(os.Getenv("CLUSTER_FQDN")),
	}, nil
}
