package main

import (
	"log"
	"net/http"
	"time"

	"github.com/yourname/aethersync/internal/engine"
	"github.com/yourname/aethersync/websocket"
)

const (
	tickRate         = 30 * time.Millisecond
	worldWidth       = 2000
	worldHeight      = 2000
	quadtreeCapacity = 4
)

func main() {
	registry := engine.NewRegistry()
	hub := websocket.NewHub(registry)
	go hub.Run()

	worldBoundary := engine.NewAABB(worldWidth/2, worldHeight/2, worldWidth/2)
	gameEngine := engine.NewGameEngine(registry, hub, worldBoundary, quadtreeCapacity)
	gameEngine.Start(tickRate)
	defer gameEngine.Stop()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, registry, w, r)
	})

	log.Println("gateway listening on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
