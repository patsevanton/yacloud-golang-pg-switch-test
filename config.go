package main

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	PGUser               string
	PGPassword           string
	PGDatabase           string
	ClusterFQDN          string
	PGSSLMode            string
	PGTargetSessionAttrs string
}

func LoadConfig() (*Config, error) {
	requiredVars := map[string]string{
		"CLUSTER_FQDN":           "переменная CLUSTER_FQDN не задана",
		"PG_SSLMODE":             "переменная PG_SSLMODE не задана",
		"PG_TARGET_SESSION_ATTRS": "переменная PG_TARGET_SESSION_ATTRS не задана",
	}

	for env, msg := range requiredVars {
		if os.Getenv(env) == "" {
			return nil, fmt.Errorf(msg)
		}
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
