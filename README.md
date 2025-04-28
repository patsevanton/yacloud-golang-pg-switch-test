# yacloud-golang-pg-switch-test

Простой код на Go для тестирования и сравнения поведения кластера PostgreSQL 
в [Yandex Cloud Managed Database](https://cloud.yandex.ru/services/managed-postgresql) при переключении мастера.

Утилита подключается к:
- **FQDN кластера** (`c-<cluster_id>.rw.mdb.yandexcloud.net`) — всегда указывает на **текущий мастер**
- Отдельным **инстансам PostgreSQL** (например, `rc1a-xxx.mdb.yandexcloud.net` и др.)

Программа фиксирует роли, доступность и стабильность соединений до, во время и после переключения мастера.

---

## 🔧 Возможности

- Подключение к FQDN и отдельным хостам PostgreSQL
- Определение текущей роли (`pg_is_in_recovery()`)
- Сравнение ролей между FQDN и прямыми хостами
- Журналирование для анализа работы отказоустойчивости

---

## 🗂 Структура проекта

```
yacloud-golang-pg-switch-test/
├── main.go        // Главная точка входа в программу. Здесь запускается логика подключения, проверок и вывода результатов.
├── config.go      // Содержит параметры конфигурации: список хостов, креды для подключения к БД, порт и другие настройки.
├── dbclient.go    // Отвечает за создание и управление соединениями с PostgreSQL. Подключается к FQDN и отдельным хостам.
├── checker.go     // Содержит функции, которые определяют роль узлов (master/replica), сравнивают их и выводят информацию.
└── README.md
```

---

## 🚀 Быстрый старт

### 1. Клонируйте репозиторий

```bash
git clone https://github.com/your-org/yacloud-golang-pg-switch-test.git
cd yacloud-golang-pg-switch-test
```

### 2. Настройте конфигурацию

Создайте `.env` или укажите параметры в `config.go`:

```env
PG_USER=your_username
PG_PASSWORD=your_password
PG_DB=your_database
CLUSTER_FQDN=c-<cluster_id>.rw.mdb.yandexcloud.net
HOSTS=rc1a-xxx.mdb.yandexcloud.net,rc1a-yyy.mdb.yandexcloud.net,rc1a-zzz.mdb.yandexcloud.net
```

### 3. Запустите утилиту

```bash
make build
make run
```

---

## 🧪 Что делает программа

1. Подключается к `CLUSTER_FQDN` и определяет мастера.
2. Подключается к каждому из `HOSTS` и проверяет их роли.
3. Сравнивает результаты — особенно полезно при тестах failover'а.
4. Выводит логи в консоль (можно доработать под Prometheus/файлы и т.п.)

---

## 📋 Пример вывода

```
[FQDN КЛАСТЕРА] Подключено: c-abcde.rw.mdb.yandexcloud.net | Роль: master
[ХОСТ] rc1a-xxx.mdb.yandexcloud.net | Роль: replica
[ХОСТ] rc1a-yyy.mdb.yandexcloud.net | Роль: master
[ХОСТ] rc1a-zzz.mdb.yandexcloud.net | Роль: replica
✅ Мастер соответствует ожидаемому
```

---

## 📌 Требования

- Go 1.18+
- Драйвер PostgreSQL: `github.com/jackc/pgx`

---

## 🤝 Контрибьютинг

Пулл-реквесты приветствуются! Открывайте issue или предлагайте улучшения.
