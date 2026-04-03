package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/digitalghost404/inkandbone/internal/ai"
	"github.com/digitalghost404/inkandbone/internal/api"
	"github.com/digitalghost404/inkandbone/internal/db"
	mcpserver "github.com/digitalghost404/inkandbone/internal/mcp"
	ttrpgweb "github.com/digitalghost404/inkandbone/web"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("home dir: %v", err)
	}
	dbPath := filepath.Join(home, ".ttrpg", "ttrpg.db")
	dataDir := filepath.Join(home, ".ttrpg")

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	var aiClient ai.Completer
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		aiClient = ai.NewClient(key)
		log.Println("AI features enabled")
	}

	httpServer := api.NewServer(database, dataDir, aiClient)

	distFS, err := fs.Sub(ttrpgweb.Static, "dist")
	if err != nil {
		log.Fatalf("embed sub: %v", err)
	}
	httpServer.RegisterStatic(http.FS(distFS))

	go func() {
		log.Println("HTTP server listening on :7432")
		if err := httpServer.ListenAndServe(":7432"); err != nil {
			log.Printf("HTTP server stopped: %v", err)
		}
	}()

	mcpSrv := mcpserver.New(database, httpServer.Bus()) // aiClient added in Task 5
	if err := mcpSrv.Start(); err != nil {
		log.Fatalf("MCP server: %v", err)
	}
}
