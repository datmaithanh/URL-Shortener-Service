package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/datmaithanh/URL-Shortener-Service/api"
	"github.com/datmaithanh/URL-Shortener-Service/config"
	db "github.com/datmaithanh/URL-Shortener-Service/db/sqlc"
	_ "github.com/lib/pq"
)

func main() {
	config.LoadConfig()
	conn, err := sql.Open(config.DB_DRIVER, config.DBSource)
	if err != nil {
		log.Fatalf("cannot connect to db: %s", err)
	}

	store := db.NewStore(conn)
	runGinServer(store)
}

func runGinServer(store db.Store) {
	server, err := api.NewServer(store)
	if err != nil {
		fmt.Printf("Cannot run server: %s", err)
	}
	err = server.Start(config.SERVER_ADDRESS)
	if err != nil {
		log.Fatalf("cannot start server: %s", err)
	}
}
