package main

import (
	"fmt"
	"net"
)

func CheckClusterFQDN(cfg *Config) {
	fmt.Println("проверка через cname:")

	fqdnIPs, err := net.LookupIP(cfg.ClusterFQDN)
	if err != nil {
		return
	}

	cname, err := net.LookupCNAME(cfg.ClusterFQDN)
	if err == nil && cname != cfg.ClusterFQDN {
		fmt.Printf("%s cname на хост %s.\n", cfg.ClusterFQDN, cname)
	}

	pool, _, err := ConnectToHost(cfg, cfg.ClusterFQDN)
	if err != nil {
		return
	}
	defer pool.Close()

	role, err := GetRole(pool)
	if err != nil {
		return
	}

	if len(fqdnIPs) == 0 {
		return
	}

	fmt.Printf("%s: %s(%s)\n", role, cfg.ClusterFQDN, fqdnIPs[0].String())
}
