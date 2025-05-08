package main

import (
	"os"
	"strings"
	"fmt"
)

type Config struct {
	PGUser                string
	PGPassword            string
	PGDatabase            string
	ClusterFQDN           string
	PGSSLMode             string
	PGTargetSessionAttrs  string
}

func LoadConfig() (*Config, error) {
	if os.Getenv("CLUSTER_FQDN") == "" {
		return nil, fmt.Errorf("переменная CLUSTER_FQDN не задана")
	}
	if os.Getenv("PG_SSLMODE") == "" {
		return nil, fmt.Errorf("переменная PG_SSLMODE не задана")
	}
	if os.Getenv("PG_TARGET_SESSION_ATTRS") == "" {
		return nil, fmt.Errorf("переменная PG_TARGET_SESSION_ATTRS не задана")
	}

	return &Config{
		PGUser:               os.Getenv("PG_USER"),
		PGPassword:           os.Getenv("PG_PASSWORD"),
		PGDatabase:           os.Getenv("PG_DB"),
		ClusterFQDN:          strings.TrimSpace(os.Getenv("CLUSTER_FQDN")),
		PGSSLMode:            os.Getenv("PG_SSLMODE"),
		PGTargetSessionAttrs: os.Getenv("PG_TARGET_SESSION_ATTRS"),
	}, nil
}
