package main

import (
	"log"

	"shorty/internal/app"
	"shorty/internal/config"
	"shorty/internal/db"
)

func main() {
	cfg := config.Load()

	gdb, err := db.OpenSQLite("shorty.db")
	if err != nil {
		log.Fatal("failed to open db:", err)
	}

	a := app.New(cfg, gdb)
	log.Fatal(a.Run(":8080"))
}
