package main

import (
	"github.com/nnxmxni/gophersocial/internals/db"
	"github.com/nnxmxni/gophersocial/internals/env"
	"github.com/nnxmxni/gophersocial/internals/store"
	"log"
)

func main() {

	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		dbConfig: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/gophersocial?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
	}

	database, err := db.New(
		cfg.dbConfig.addr,
		cfg.dbConfig.maxOpenConns,
		cfg.dbConfig.maxIdleConns,
		cfg.dbConfig.maxIdleTime,
	)
	if err != nil {
		log.Panic(err)
	}

	defer database.Close()

	log.Println("Database connection pool established")

	newStorage := store.NewStorage(database)

	app := &application{
		config: cfg,
		store:  newStorage,
	}

	mux := app.mount()

	log.Fatal(app.run(mux))
}
