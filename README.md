# Тестирование

Для тестов требуется:

* запущенный Docker контейнер с Postgres
* запущенный Docker контейнер с Mongo

## 1. Установка переменной окружения с паролем для Postgres

```console
export POSTGRES_PASSWORD='some_pass'
```

## 2. Запустить контейнеры

```console
chmod +x cmd/run_postgres.sh
chmod +x cmd/run_mongo.sh
./cmd/run_postgres.sh
./cmd/run_mongo.sh
```

## 3. Запуск тестов

```console
go test -v ./...
```
