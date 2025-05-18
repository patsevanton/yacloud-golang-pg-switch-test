resource "yandex_mdb_postgresql_cluster" "my_cluster" {
  name        = "ha"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config {
    version = 16
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.baz.id
    assign_public_ip = true
  }

  host {
    zone      = "ru-central1-b"
    subnet_id = yandex_vpc_subnet.foo.id
    assign_public_ip = true
  }

  host {
    zone      = "ru-central1-d"
    subnet_id = yandex_vpc_subnet.bar.id
    assign_public_ip = true
  }

  timeouts {
    create = "1h30m" # Полтора часа
    update = "2h"    # 2 часа
    delete = "1h"   # 30 минут
  }
}

resource "yandex_mdb_postgresql_user" "my_user" {
  cluster_id = yandex_mdb_postgresql_cluster.my_cluster.id
  name       = "test"
  password   = "SuperSecurePassw0rd"
  conn_limit = 50
}

resource "yandex_mdb_postgresql_database" "my_db" {
  cluster_id = yandex_mdb_postgresql_cluster.my_cluster.id
  name       = "testdb"
  owner      = yandex_mdb_postgresql_user.my_user.name

}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "baz" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}

output "cluster_fqdn" {
  description = "FQDN for the PostgreSQL cluster"
  value       = "c-${yandex_mdb_postgresql_cluster.my_cluster.id}.rw.mdb.yandexcloud.net"
}

output "hosts_fqdns" {
  description = "FQDNs of all cluster hosts"
  value       = [
    for host in yandex_mdb_postgresql_cluster.my_cluster.host : host.fqdn
  ]
}

resource "local_file" "env_file" {
  filename = "${path.module}/.env"
  content  = <<EOT
PG_HOST=c-${yandex_mdb_postgresql_cluster.my_cluster.id}.rw.mdb.yandexcloud.net
PG_PORT=6432
PG_USER=${yandex_mdb_postgresql_user.my_user.name}
PG_PASSWORD=${yandex_mdb_postgresql_user.my_user.password}
PG_DB=${yandex_mdb_postgresql_database.my_db.name}
POOL_MAX_CONNS=2
POOL_MIN_CONNS=10
POOL_MAX_CONN_LIFETIME=1h
POOL_MAX_CONN_IDLE_TIME=30m
DEFAULT_QUERY_EXEC_MODE=simple_protocol
EOT
}
