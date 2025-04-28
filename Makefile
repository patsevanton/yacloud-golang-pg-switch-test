.PHONY: build run tidy

build:
	go build -o switch-checker check_hosts.go config.go connect_to_host.go get_role.go hide_password_in_dsn.go main.go check_cluster_fqdn.go

run: tidy build
	./switch-checker

tidy:
	go mod tidy
