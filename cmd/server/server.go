package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"GoNews/pkg/api"
	"GoNews/pkg/storage"
	"GoNews/pkg/storage/memdb"
	"GoNews/pkg/storage/mongo"
	"GoNews/pkg/storage/postgres"
)

// Сервер GoNews.
type server struct {
	db  storage.Interface
	api *api.API
}

func main() {
	// Создаём объект сервера.
	var (
		srv    server
		dbType string
	)

	flag.StringVar(&dbType, "db", "memdb", "Specify database for the application. Available: memdb, postgres, mongo")
	flag.Parse()

	switch dbType {
	case "memdb":
		// Создаём объекты баз данных.
		//
		// БД в памяти.
		srv.db = memdb.New()

	case "postgres":
		// Реляционная БД PostgreSQL.
		conf := postgres.Config{
			User:     "postgres",
			Password: os.Getenv("POSTGRES_PASSWORD"),
			Host:     "localhost",
			Port:     "5433",
			DBName:   "gonews",
		}
		db, err := postgres.New(conf.ConString())
		if err != nil {
			log.Fatal(err)
		}
		srv.db = db

	case "mongo":
		// Документная БД MongoDB.
		conf := mongo.Config{
			Host:   "localhost",
			Port:   "27018",
			DBName: "gonews",
		}
		db, err := mongo.New(conf)
		if err != nil {
			log.Fatal(err)
		}
		srv.db = db

	default:
		log.Fatal("Invalid DB type specified")
	}

	// Создаём объект API и регистрируем обработчики.
	srv.api = api.New(srv.db)

	// Запускаем веб-сервер на порту 8080 на всех интерфейсах.
	// Предаём серверу маршрутизатор запросов,
	// поэтому сервер будет все запросы отправлять на маршрутизатор.
	// Маршрутизатор будет выбирать нужный обработчик.
	http.ListenAndServe(":8080", srv.api.Router())
}
