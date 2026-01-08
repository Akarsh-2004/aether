package main

import (
	"flag"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	aetherpb "github.com/yourname/aethersync/proto"
	"google.golang.org/protobuf/proto"
)

var (
	addr       = flag.String("addr", "localhost:8080", "http service address")
	numClients = flag.Int("clients", 100, "number of clients to simulate")
	duration   = flag.Int("duration", 30, "duration of test in seconds")
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < *numClients; i++ {
		wg.Add(1)
		go runClient(i, u.String(), &wg)
		time.Sleep(time.Millisecond * 10) // Stagger connections
	}

	go func() {
		time.Sleep(time.Duration(*duration) * time.Second)
		log.Printf("Test duration reached, stopping...")
		close(interrupt)
	}()

	<-interrupt
	log.Printf("Test finished. Total time: %v", time.Since(start))
}

func runClient(id int, urlStr string, wg *sync.WaitGroup) {
	defer wg.Done()

	c, _, err := websocket.DefaultDialer.Dial(urlStr, nil)
	if err != nil {
		log.Printf("[Client %d] dial error: %v", id, err)
		return
	}
	defer c.Close()

	done := make(chan struct{})

	// Read loop
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				return
			}

			var snapshot aetherpb.WorldSnapshot
			if err := proto.Unmarshal(message, &snapshot); err != nil {
				log.Printf("[Client %d] unmarshal error: %v", id, err)
				continue
			}
			// In a real load test, we could track latency here if we had timestamps in the snapshot
		}
	}()

	ticker := time.NewTicker(time.Millisecond * 100) // Send input every 100ms
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// Send random velocity
			input := &aetherpb.ClientInput{
				VelocityX: rand.Float32()*2 - 1,
				VelocityY: rand.Float32()*2 - 1,
			}
			data, _ := proto.Marshal(input)
			err := c.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				log.Printf("[Client %d] write error: %v", id, err)
				return
			}
		case <-time.After(time.Duration(*duration) * time.Second):
			return
		}
	}
}
