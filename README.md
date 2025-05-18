## Всегда используйте target_session_attrs=read-write (или primary) при подключении к кластеру PostgreSQL

При работе с кластерами PostgreSQL, особенно в конфигурациях с высокой доступностью (High Availability, HA), 
разработчики часто сталкиваются с ошибками типа "cannot execute INSERT in a read-only transaction". Эти ошибки 
возникают, когда приложение пытается выполнить операцию записи на узел, который в данный момент является репликой 
(read-only). Особенно остро эта проблема проявляется в моменты переключения мастера: пул соединений вашего 
приложения может все еще содержать коннекты к бывшему мастеру, ставшему репликой, или же балансировщик может 
направить новый запрос на запись к реплике.

К счастью, драйверы PostgreSQL предоставляют элегантное решение этой проблемы – параметр строки подключения 
`target_session_attrs`. В версии PostgreSQL 14 были добавлены новые значения для target_session_attrs: read-only, primary, 
standby и prefer-standby. Этот параметр позволяет указать, какого типа сессию ожидает ваше приложение. Наиболее полезным 
значением для приложений, выполняющих операции чтения и записи, является `read-write`. В этой статье мы подробно разберем, 
почему это так важно, продемонстрируем проблему на практике и покажем, как `target_session_attrs=read-write` спасает ситуацию.
Использование `target_session_attrs=primary`, включая его специфические отличия от `read-write`, будет подробно рассмотрено далее в статье.

### Подготовка тестового окружения: кластер PostgreSQL в Яндекс.Облаке

Для наглядной демонстрации мы развернем отказоустойчивый кластер PostgreSQL в Яндекс.Облаке. Это позволит нам легко 
симулировать переключение мастера и наблюдать за поведением приложения. Для управления инфраструктурой мы будем 
использовать Terraform.

Ниже приведен код Terraform (`main.tf`), который описывает создание кластера PostgreSQL с тремя хостами в разных зонах 
доступности, а также необходимой сети, подсетей, пользователя и базы данных. Этот кластер будет сконфигурирован для 
высокой доступности.

```terraform
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
  password   = "" # укажите здесь пароль к PostgreSQL
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
POOL_MIN_CONNS=2
POOL_MAX_CONN_LIFETIME=1h
POOL_MAX_CONN_IDLE_TIME=30m
DEFAULT_QUERY_EXEC_MODE=simple_protocol
TARGET_SESSION_ATTRS=any
EOT
}

```

Этот Terraform-код создаст:
*   VPC (виртуальное частное облако) и три подсети в разных зонах доступности (`ru-central1-a`, `ru-central1-b`, `ru-central1-d`).
*   Кластер Managed Service for PostgreSQL с тремя хостами (один мастер и две реплики), распределенными по этим зонам. 
*   Мы используем версию PostgreSQL 16 и небольшие ресурсы (`s2.micro`, 16 ГБ SSD) для демонстрационных целей.
*   Пользователя `test` с паролем и базу данных `testdb`.
*   Также, для удобства, будет сгенерирован файл `.env` со всеми необходимыми параметрами подключения к кластеру. 
*   Обратите внимание, что в этом файле параметр `TARGET_SESSION_ATTRS` по умолчанию установлен в `any`. Это важно для 
*   нашего первого теста.

После применения этого Terraform-кода (`terraform apply`) у нас будет готовый к работе кластер PostgreSQL.

### Выбор языка для тестирования: Go

Для написания тестового приложения, которое будет подключаться к нашему кластеру и выполнять операции записи, мы выбрали 
язык Go (Golang). Go является достаточно популярным языком для разработки бэкенд-сервисов, обладает отличной стандартной 
библиотекой, мощными инструментами для работы с конкурентностью и имеет зрелые драйверы для работы с PostgreSQL, 
такие как `pgx`.

Наше тестовое приложение (`main.go`) будет в бесконечном цикле пытаться вставить данные в таблицу, логируя результат 
каждой попытки, включая информацию о пуле соединений и узле, к которому было выполнено подключение.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: no .env file found, trying to read env from system")
	}

	ctx := context.TODO()

	connString := buildConnStringFromEnv()
	log.Println(connString)

	db, pool, err := GetDB(ctx, connString)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected!!")

	if _, err := db.Exec(`create table if not exists test (
	   id bigint GENERATED BY DEFAULT AS IDENTITY primary key,
	   created_at timestamptz not null
	 )`); err != nil {
		log.Fatalf("unable to create table: %v\n", err)
	}

	for {
		result, err := db.Exec(
			`insert into "test" ("created_at") values ($1)`, time.Now())

		if err != nil {
			connInfo := GetConnectionInfo(ctx, pool)
			pgErr, ok := err.(*pgconn.PgError)
			if ok {
				log.Printf("DB Error [%s] %s (Code: %s): %s", connInfo, pgErr.Message, pgErr.Code, err)
			} else {
				log.Printf("DB Error [%s]: %s", connInfo, err)
			}
			time.Sleep(1 * time.Second)
			continue
		}

		connInfo := GetConnectionInfo(ctx, pool)
		rowsAffected, _ := result.RowsAffected()
		log.Printf("Successful [%s]: rows affected: %d", connInfo, rowsAffected)
		time.Sleep(1 * time.Second)
	}
}

func buildConnStringFromEnv() string {
	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASSWORD")
	host := os.Getenv("PG_HOST")
	port := os.Getenv("PG_PORT")
	db := os.Getenv("PG_DB")

	poolMaxConns := os.Getenv("POOL_MAX_CONNS")
	poolMinConns := os.Getenv("POOL_MIN_CONNS")
	poolMaxConnLifetime := os.Getenv("POOL_MAX_CONN_LIFETIME")
	poolMaxConnIdleTime := os.Getenv("POOL_MAX_CONN_IDLE_TIME")
	defaultQueryExecMode := os.Getenv("DEFAULT_QUERY_EXEC_MODE")
	targetSessionAttrs := os.Getenv("TARGET_SESSION_ATTRS")

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?pool_max_conns=%s&pool_min_conns=%s&pool_max_conn_lifetime=%s" +
		"&pool_max_conn_idle_time=%s&default_query_exec_mode=%s&target_session_attrs=%s",
		user, password, host, port, db, poolMaxConns, poolMinConns, poolMaxConnLifetime,
		poolMaxConnIdleTime, defaultQueryExecMode, targetSessionAttrs,
	)

	return connString
}

func GetConnectionInfo(ctx context.Context, pool *pgxpool.Pool) string {
	var connInfo strings.Builder
	stat := pool.Stat()
	connInfo.WriteString(fmt.Sprintf("Pool total: %d, acquired: %d, idle: %d | ",
		stat.TotalConns(), stat.AcquiredConns(), stat.IdleConns()))

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return connInfo.String() + "Failed to acquire connection"
	}
	defer conn.Release()

	remoteAddr := conn.Conn().PgConn().Conn().RemoteAddr().String()
	host, _, _ := net.SplitHostPort(remoteAddr)

	var readOnly bool
	err = conn.QueryRow(ctx, "SELECT pg_is_in_recovery()").Scan(&readOnly)
	if err != nil {
		connInfo.WriteString(fmt.Sprintf("IP: %s, Error getting read-only status", host))
		return connInfo.String()
	}

	serverType := "Master"
	if readOnly {
		serverType = "Replica"
	}

	connInfo.WriteString(fmt.Sprintf("IP: %s, Type: %s", host, serverType))
	return connInfo.String()
}

func GetDB(ctx context.Context, uri string) (*sqlx.DB, *pgxpool.Pool, error) {
	DB, pool, err := PgxCreateDB(ctx, uri)
	if err != nil {
		return nil, nil, err
	}
	return DB, pool, nil
}

func PgxCreateDB(ctx context.Context, uri string) (*sqlx.DB, *pgxpool.Pool, error) {
	connConfig, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse connection config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	pgxdb := stdlib.OpenDBFromPool(pool)
	return sqlx.NewDb(pgxdb, "pgx"), pool, nil
}
```

Ключевые моменты в коде `main.go`:
1.  **Сборка строки подключения**: Функция `buildConnStringFromEnv` читает параметры из переменных окружения (которые мы получаем из `.env` файла, сгенерированного Terraform) и формирует строку подключения. Важно, что `target_session_attrs` также берется из окружения.
2.  **Получение информации о соединении**: Функция `GetConnectionInfo` перед каждой операцией (или в случае ошибки) запрашивает у пула соединений `pgxpool.Pool` информацию о текущем состоянии пула. Затем она берет одно соединение из пула, определяет его удаленный IP-адрес и запрашивает у PostgreSQL, является ли текущий узел репликой (`SELECT pg_is_in_recovery()`). Это позволяет нам точно знать, на какой узел (мастер или реплика) пошла операция.
3.  **Цикл вставок**: В `main` функции приложение создает таблицу `test`, если она не существует, а затем входит в бесконечный цикл, каждую секунду пытаясь вставить новую запись с текущим временем. Логируется либо успешная вставка, либо ошибка, всегда с указанием информации о соединении.

Для сборки и запуска приложения мы будем использовать простой `Makefile`:

```makefile
.PHONY: build run tidy

run:
	go mod tidy
	go build -o switch-checker *.go
	./switch-checker

```

Команда `make run` сначала выполнит `go mod tidy`, затем соберет исполняемый файл `switch-checker` и запустит его.

### Тестирование переключения мастера БЕЗ `target_session_attrs` (или с `target_session_attrs=any`)

Итак, у нас развернут кластер PostgreSQL в Яндекс.Облаке, и Terraform сгенерировал `.env` файл, в котором 
`TARGET_SESSION_ATTRS` установлен в `any`. Это значение означает, что драйверу разрешено подключаться к любому 
доступному хосту кластера, независимо от его роли (мастер или реплика).

Запустим наше Go-приложение:
1.  Убедитесь, что вы находитесь в директории с файлами `main.go`, `main.tf` и `Makefile`.
2.  Если вы еще не развернули инфраструктуру, выполните `terraform init` и `terraform apply`.
3.  После успешного создания кластера и файла `.env`, выполните `make run`.

Приложение начнет выполнять INSERT-запросы. В нормальном состоянии все запросы будут идти на мастер-узел. Теперь давайте 
сымитируем сбой мастера или плановое переключение. В консоли управления Яндекс.Облака для вашего кластера Managed Service 
for PostgreSQL вы можете инициировать операцию переключения мастера (Start Failover).

После того как переключение мастера произойдет, мы начнем наблюдать в логах нашего приложения следующую картину:

```
2025/05/18 12:45:10 Connected!!
2025/05/18 12:45:10 postgres://test:пароль@c-xxxxxxxxx.rw.mdb.yandexcloud.net:6432/testdb?pool_max_conns=2&pool_min_conns=2&pool_max_conn_lifetime=1h&pool_max_conn_idle_time=30m&default_query_exec_mode=simple_protocol&target_session_attrs=any
... (успешные вставки)
2025/05/18 12:46:16 Successful [Pool total: 2, acquired: 0, idle: 2 | IP: 62.84.123.60, Type: Master]: rows affected: 1
// Начинается переключение мастера, или приложение подключилось к реплике
2025/05/18 12:46:17 DB Error [Pool total: 2, acquired: 0, idle: 2 | IP: 89.169.156.108, Type: Replica] cannot execute INSERT in a read-only transaction (Code: 25006): ERROR: cannot execute INSERT in a read-only transaction (SQLSTATE 25006)
2025/05/18 12:46:17 DB Error [Pool total: 2, acquired: 0, idle: 2 | IP: 89.169.156.108, Type: Replica] cannot execute INSERT in a read-only transaction (Code: 25006): ERROR: cannot execute INSERT in a read-only transaction (SQLSTATE 25006)
// ... множество таких ошибок
2025/05/18 12:46:17 DB Error [Pool total: 2, acquired: 0, idle: 2 | IP: 89.169.156.108, Type: Replica] cannot execute INSERT in a read-only transaction (Code: 25006): ERROR: cannot execute INSERT in a read-only transaction (SQLSTATE 25006)
// Пул соединений мог обновить информацию и подключиться к новому мастеру, либо другой коннект из пула был к мастеру
2025/05/18 12:46:18 Successful [Pool total: 2, acquired: 0, idle: 2 | IP: 62.84.123.60, Type: Master]: rows affected: 1
// Но через некоторое время ошибки могут повториться, если соединения к репликам остаются в пуле
2025/05/18 12:47:19 DB Error [Pool total: 2, acquired: 0, idle: 2 | IP: 89.169.156.108, Type: Replica] cannot execute INSERT in a read-only transaction (Code: 25006): ERROR: cannot execute INSERT in a read-only transaction (SQLSTATE 25006)
```

**Что мы видим в логах и почему это происходит?**

1.  **Определение типа хоста и IP**: Наш Go-код в функции `GetConnectionInfo` выполняет запрос `SELECT pg_is_in_recovery()`. Если он возвращает `true`, узел является репликой, иначе – мастером. Мы также получаем IP-адрес узла, к которому подключено данное соединение.
2.  **Подтверждение роли реплики**: Когда мы видим в логе `IP: 89.169.156.108, Type: Replica`, это означает, что наше приложение установило соединение с хостом `89.169.156.108`, и этот хост подтвердил, что он находится в режиме восстановления (т.е. является репликой).
3.  **Проблема с пулом соединений**: Библиотека `pgxpool` (и большинство других пулов соединений) поддерживает некоторое количество открытых соединений для повышения производительности. При `target_session_attrs=any`, пул может устанавливать соединения как с мастером, так и с репликами через общий FQDN кластера (например, `c-xxxxxxxxx.rw.mdb.yandexcloud.net`), который обычно указывает на текущий мастер, но при определенных обстоятельствах или при прямом подключении к FQDN реплик (если бы мы их использовали) может направить на реплику.
    Более того, даже если изначально все соединения в пуле были к мастеру, после переключения мастера эти "старые" соединения теперь ведут к узлу, который стал репликой. Приложение, беря такое соединение из пула для операции `INSERT`, сталкивается с ошибкой `cannot execute INSERT in a read-only transaction (Code: 25006)`.
4.  **Чередование ошибок и успехов**: Как видно из логов, предоставленных в примере, приложение может какое-то время получать ошибки, а затем успешно выполнить вставку. Это происходит потому, что в пуле (в нашем примере `POOL_MAX_CONNS=2`) может быть несколько соединений. Одно соединение может все еще указывать на старый мастер (теперь реплику), вызывая ошибку. Другое соединение могло быть установлено уже к новому мастеру (например, `IP: 62.84.123.60, Type: Master`), и операция через него пройдет успешно. Это создает нестабильное поведение и потерю данных (или необходимость сложной логики повторов в приложении).

Этот тест наглядно демонстрирует проблему: без явного указания на необходимость подключения к мастер-узлу для операций записи, приложение становится уязвимым к ошибкам во время переключения мастера или при неоптимальной конфигурации балансировки нагрузки на уровне кластера.

### Тестирование переключения мастера с `target_session_attrs=read-write`

Теперь давайте исправим эту ситуацию. Мы изменим параметр `TARGET_SESSION_ATTRS` в нашем файле `.env` на `read-write`. 
Это значение указывает драйверу `pgx`, что мы намерены подключаться к узлу, который способен обрабатывать как чтение, 
так и запись, то есть к мастеру.

1.  Остановите работающее Go-приложение (Ctrl+C).
2.  Откройте файл `.env` в вашем текстовом редакторе.
3.  Найдите строку `TARGET_SESSION_ATTRS=any` и измените ее на `TARGET_SESSION_ATTRS=read-write`.
    Содержимое `.env` теперь должно выглядеть примерно так (FQDN, пользователь, пароль и имя БД будут вашими):
    ```dotenv
    PG_HOST=c-xxxxxxxxx.rw.mdb.yandexcloud.net
    PG_PORT=6432
    PG_USER=test
    PG_PASSWORD=пароль
    PG_DB=testdb
    POOL_MAX_CONNS=2
    POOL_MIN_CONNS=2
    POOL_MAX_CONN_LIFETIME=1h
    POOL_MAX_CONN_IDLE_TIME=30m
    DEFAULT_QUERY_EXEC_MODE=simple_protocol
    TARGET_SESSION_ATTRS=read-write
    ```
4.  Сохраните изменения в файле `.env`.
5.  Снова запустите приложение: `make run`.

Теперь приложение будет подключаться к кластеру с инструкцией искать узел, поддерживающий чтение и запись. Драйвер `libpq` 
(который используется `pgx` под капотом при подключении к нескольким хостам или через специальный FQDN, как у Яндекс.Облака) 
будет использовать эту информацию для выбора подходящего хоста. Если текущее соединение ведет к реплике, драйвер попытается 
установить новое соединение с мастером.

Снова инициируйте переключение мастера в консоли Яндекс.Облака.

Наблюдайте за логами приложения. Вы заметите существенное отличие:

```
2025/05/18 13:10:15 Connected!!
2025/05/18 13:10:15 postgres://test:пароль@c-xxxxxxxxx.rw.mdb.yandexcloud.net:6432/testdb?pool_max_conns=2&pool_min_conns=2&pool_max_conn_lifetime=1h&pool_max_conn_idle_time=30m&default_query_exec_mode=simple_protocol&target_session_attrs=read-write
... (успешные вставки на текущий мастер)
2025/05/18 13:11:20 Successful [Pool total: 2, acquired: 0, idle: 2 | IP: 62.84.123.60, Type: Master]: rows affected: 1
// Происходит переключение мастера
// Может быть короткая пауза или несколько попыток переподключения, но ошибок "read-only transaction" быть не должно
// Драйвер и пул соединений отработают подключение к новому мастеру
2025/05/18 13:11:55 Successful [Pool total: 2, acquired: 0, idle: 2 | IP: 84.252.135.217, Type: Master]: rows affected: 1 // IP изменился на IP нового мастера
2025/05/18 13:11:56 Successful [Pool total: 2, acquired: 0, idle: 2 | IP: 84.252.135.217, Type: Master]: rows affected: 1
... (дальнейшие успешные вставки)
```

В логах вы больше не должны видеть ошибок `cannot execute INSERT in a read-only transaction`. Приложение может испытать 
короткую задержку во время фактического переключения мастера, пока драйвер устанавливает соединение с новым мастером, но 
операции записи не будут направляться на реплики. Пул соединений, работая совместно с драйвером, который учитывает 
`target_session_attrs=read-write`, будет гарантировать, что для операций используются только соединения с мастером. 
Любые соединения, которые после переключения мастера оказываются подключенными к реплике, не будут использоваться для 
операций записи или будут переустановлены.

### Особый случай: `target_session_attrs=primary` и проблема нехватки места на диске

Помимо `read-write`, еще одним важным значением для `target_session_attrs`, является `primary`. В то время как `read-write` позволяет драйверу подключаться к любому серверу, который не находится в режиме горячего резерва (`hot standby`, то есть `pg_is_in_recovery()` возвращает `false`), значение `primary` более строго указывает на необходимость подключения именно к тому серверу, который кластером идентифицирован как основной (primary). В большинстве штатных ситуаций оба этих значения приведут к подключению к одному и тому же мастер-серверу.

Однако существует важное различие в поведении, которое может проявиться в критических ситуациях, например, когда на основном сервере заканчивается дисковое пространство. Рассмотрим этот сценарий:

Предположим, на дисках PostgreSQL кластера, где хранятся данные основного сервера, не осталось свободного места.

*   При использовании `target_session_attrs=read-write`: В такой ситуации, хотя основной сервер все еще является мастером (не в `hot standby` и `pg_is_in_recovery()` для него `false`), его способность выполнять операции записи фактически утрачена из-за нехватки места. Существует вероятность, что система подключения — будь то сам `libpq` при выборе из списка хостов или логика работы специализированного FQDN балансировщика (например, `*.rw.mdb.yandexcloud.net`) — может не суметь установить или предоставить приложению сессию, которую она считает действительно готовой к записи. Например, пул соединений может испытывать трудности с получением "здорового" соединения для записи, или новые попытки подключения могут не найти узел, который система считает полностью удовлетворяющим критерию "готов к записи" из-за проблем с ресурсами. Это может привести к тому, что приложение не сможет установить соединение с мастером для выполнения операций записи, и, в зависимости от его логики обработки ошибок подключения, может даже завершить свою работу или перейти в состояние, когда операции с базой данных невозможны.

*   При использовании `target_session_attrs=primary`: Это указание является более прямолинейным и детерминированным. `libpq` будет стремиться подключиться именно к тому серверу, который в конфигурации кластера объявлен главным (primary). Даже если на этом сервере закончилось дисковое пространство, он по-прежнему сохраняет свою роль `primary`. Соединение с ним, с высокой вероятностью, будет установлено успешно. Однако, когда приложение попытается выполнить операцию записи (например, `INSERT`, `UPDATE`, `DELETE`), оно получит ошибку непосредственно от СУБД PostgreSQL, указывающую на нехватку дискового пространства. Критически важным аспектом здесь является то, что само приложение (сервис) останется в рабочем состоянии. Оно сможет корректно залогировать эту специфическую ошибку от базы данных и, возможно, продолжить обслуживать другие запросы (например, на чтение, если они обрабатываются отдельно) или адекватно информировать пользователей или другие системы о временной невозможности выполнения операций записи данных.

**Когда выбирать `target_session_attrs=primary`?**

Использование `target_session_attrs=primary` становится предпочтительным для тех сервисов, для которых критически важно оставаться в "доступном" состоянии (то есть, приложение продолжает работать) и иметь возможность точно диагностировать и обрабатывать проблемы на стороне базы данных, даже если эти проблемы (как нехватка дискового пространства) делают операции записи невозможными. Если для вашего приложения важнее получить конкретную ошибку от PostgreSQL о нехватке места и на основе этой информации принять решение (например, уведомить администраторов, временно отключить функционал записи, перейти в режим "только чтение" на уровне приложения), чем потенциально столкнуться с полной невозможностью установить "пишущую" сессию через `read-write` и, как следствие, аварийно остановиться или потерять связь с БД на уровне установления соединения, то `primary` будет более подходящим выбором.

Таким образом, `target_session_attrs=primary` обеспечивает более предсказуемое поведение при подключении к главному узлу, особенно в условиях его ресурсного истощения (например, нехватки дискового пространства). Этот параметр позволяет приложению установить соединение с основным сервером и получить от него ошибку выполнения запроса, перекладывая ответственность за обработку этой ошибки на логику приложения, но сохраняя при этом само приложение работоспособным и осведомленным о состоянии СУБД.

### Вывод

Как мы продемонстрировали, использование параметра строки подключения `target_session_attrs` со значением `read-write` 
является критически важным для обеспечения стабильной работы приложений с кластерами PostgreSQL в конфигурациях с 
высокой доступностью, особенно для гладкого прохождения переключений мастера.

**Ключевые преимущества использования `target_session_attrs=read-write`:**

1.  **Устойчивость к переключению мастера**: Приложение автоматически будет направлять запросы на запись на актуальный мастер-узел, минимизируя или полностью устраняя ошибки "read-only transaction" во время и после переключения.
2.  **Упрощение логики приложения**: Нет необходимости реализовывать сложную логику повторных попыток или обнаружения мастера на стороне приложения, так как эту задачу берет на себя драйвер СУБД.
3.  **Корректная работа с пулом соединений**: Пул соединений будет корректно обрабатывать соединения, гарантируя, что для операций записи используются только подходящие (master) коннекты.

Другие возможные значения для `target_session_attrs` включают:
*   `any` (по умолчанию, если не указано): подключаться к любому хосту. Как мы видели, это приводит к проблемам.
*   `read-only`: подключаться к хосту, который предпочтительно является репликой. Полезно для распределения нагрузки чтения.
*   `primary`: подключаться к основному серверу. Как мы рассмотрели выше, это значение имеет более строгую семантику, чем `read-write`, и предлагает иное поведение в сценариях с нехваткой ресурсов на мастере, что может быть критично для доступности самого приложения и корректной обработки ошибок.
*   `standby`: подключаться к резервному серверу (синоним `read-only`).
*   `prefer-standby`: подключаться к резервному серверу, если доступен, иначе к основному.

Для большинства приложений, выполняющих как чтение, так и запись, `target_session_attrs=read-write` остается очень надежным выбором для обеспечения отказоустойчивости при переключениях мастера. Однако, если для вашего сервиса крайне важна непрерывная работа самого приложения с возможностью получать и обрабатывать ошибки от базы данных в ситуациях, когда мастер-узел испытывает проблемы с ресурсами (например, нехватка дискового пространства), то `target_session_attrs=primary` может оказаться более предпочтительным. Выбор между `read-write` и `primary` зависит от конкретных требований вашего приложения к поведению в таких граничных условиях.

Поэтому, если вы работаете с кластером PostgreSQL, **тщательно выбирайте между `target_session_attrs=read-write` и `target_session_attrs=primary`** в строке подключения вашего приложения, понимая их нюансы. Это небольшое, но важное решение может значительно повысить надежность и предсказуемость вашей системы.

Исходный код: https://github.com/patsevanton/yacloud-golang-pg-switch-test

## Telegram канал:
Подписывайтесь на мой telegram канал https://t.me/notes_devops_engineer
