package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/digitalghost404/inkandbone/internal/api"
	"github.com/digitalghost404/inkandbone/internal/db"
	ttrpgweb "github.com/digitalghost404/inkandbone/web"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("home dir: %v", err)
	}
	dbPath := filepath.Join(home, ".ttrpg", "ttrpg.db")

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	server := api.NewServer(database)

	// Serve embedded React SPA for all non-API routes
	distFS, err := fs.Sub(ttrpgweb.Static, "dist")
	if err != nil {
		log.Fatalf("embed sub: %v", err)
	}
	server.RegisterStatic(http.FS(distFS))

	log.Println("TTRPG Companion running at http://localhost:7432")
	if err := server.ListenAndServe(":7432"); err != nil {
		log.Fatalf("server: %v", err)
	}
}
