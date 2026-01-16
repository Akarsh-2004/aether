package gateway

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/akarsh-2004/aether/internal/config"
	"github.com/akarsh-2004/aether/internal/engine"
	"github.com/akarsh-2004/aether/internal/protocol"
	"github.com/akarsh-2004/aether/proto"
	"go.uber.org/zap"
)

var (
	ErrConnectionClosed = errors.New("connection closed")
	ErrMessageTooLarge  = errors.New("message too large")
)

type WebSocketGateway struct {
	config    config.GatewayConfig
	engine    *engine.SpatialEngine
	codec     *protocol.Codec
	logger    *zap.Logger
	upgrader  websocket.Upgrader
	clients   sync.Map // map[string]*Client
	shutdown  chan struct{}
	wg        sync.WaitGroup
}

type Client struct {
	id         string
	conn       *websocket.Conn
	sendChan   chan []byte
	closeChan  chan struct{}
	entityID   uint32
	lastSeq    uint64
	mu         sync.RWMutex
}

func NewWebSocketGateway(cfg config.GatewayConfig, eng *engine.SpatialEngine, logger *zap.Logger) *WebSocketGateway {
	return &WebSocketGateway{
		config:   cfg,
		engine:   eng,
		codec:    protocol.NewCodec(),
		logger:   logger,
		shutdown: make(chan struct{}),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  cfg.ReadBufferSize,
			WriteBufferSize: cfg.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true // TODO: Implement proper origin checking
			},
			EnableCompression: cfg.EnableCompression,
		},
	}
}

func (g *WebSocketGateway) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", g.handleWebSocket)

	server := &http.Server{
		Addr:    g.config.BindAddr,
		Handler: mux,
	}

	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		g.logger.Info("WebSocket gateway starting", zap.String("addr", g.config.BindAddr))
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			g.logger.Error("WebSocket server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	g.logger.Info("WebSocket gateway shutting down")
	
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	return server.Shutdown(shutdownCtx)
}

func (g *WebSocketGateway) Shutdown(ctx context.Context) error {
	close(g.shutdown)
	
	g.clients.Range(func(key, value interface{}) bool {
		if client, ok := value.(*Client); ok {
			client.Close()
		}
		return true
	})
	
	done := make(chan struct{})
	go func() {
		g.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (g *WebSocketGateway) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := g.upgrader.Upgrade(w, r, nil)
	if err != nil {
		g.logger.Error("Failed to upgrade WebSocket connection", zap.Error(err))
		return
	}

	clientID := g.generateClientID()
	client := &Client{
		id:        clientID,
		conn:      conn,
		sendChan:  make(chan []byte, 256), // Buffered channel for non-blocking sends
		closeChan: make(chan struct{}),
	}

	g.clients.Store(clientID, client)
	g.logger.Info("Client connected", zap.String("client_id", clientID))

	g.wg.Add(2)
	go g.readPump(client)
	go g.writePump(client)
}

func (g *WebSocketGateway) readPump(client *Client) {
	defer g.wg.Done()
	defer func() {
		client.Close()
		g.clients.Delete(client.id)
		g.logger.Info("Client disconnected", zap.String("client_id", client.id))
		
		if client.entityID != 0 {
			g.engine.RemoveEntity(client.entityID)
		}
	}()

	client.conn.SetReadLimit(int64(g.config.MaxMessageSize))
	client.conn.SetReadDeadline(time.Now().Add(time.Duration(g.config.PongWait) * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(time.Duration(g.config.PongWait) * time.Second))
		return nil
	})

	for {
		select {
		case <-g.shutdown:
			return
		case <-client.closeChan:
			return
		default:
		}

		messageType, data, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				g.logger.Error("WebSocket read error", zap.String("client_id", client.id), zap.Error(err))
			}
			return
		}

		if messageType != websocket.BinaryMessage {
			g.logger.Warn("Received non-binary message", zap.String("client_id", client.id))
			continue
		}

		if len(data) > g.config.MaxMessageSize {
			g.logger.Warn("Message too large", zap.String("client_id", client.id), zap.Int("size", len(data)))
			continue
		}

		g.handleMessage(client, data)
	}
}

func (g *WebSocketGateway) writePump(client *Client) {
	defer g.wg.Done()
	defer client.Close()

	ticker := time.NewTicker(time.Duration(g.config.PingPeriod) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-g.shutdown:
			return
		case <-client.closeChan:
			return
		case message, ok := <-client.sendChan:
			client.conn.SetWriteDeadline(time.Now().Add(time.Duration(g.config.WriteWait) * time.Second))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				g.logger.Error("Failed to write message", zap.String("client_id", client.id), zap.Error(err))
				return
			}

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(time.Duration(g.config.WriteWait) * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (g *WebSocketGateway) handleMessage(client *Client, data []byte) {
	msg, err := g.codec.Decode(data)
	if err != nil {
		g.logger.Error("Failed to decode message", zap.String("client_id", client.id), zap.Error(err))
		return
	}

	if err := g.codec.ValidateMessage(msg); err != nil {
		g.logger.Error("Invalid message", zap.String("client_id", client.id), zap.Error(err))
		return
	}

	switch msg.Type {
	case proto.MessageType_MOVEMENT_DELTA:
		g.handleMovementDelta(client, msg.MovementDelta)
		
	case proto.MessageType_SPAWN_REQUEST:
		g.handleSpawnRequest(client, msg.SpawnRequest)
		
	case proto.MessageType_HEARTBEAT:
		g.handleHeartbeat(client, msg.Heartbeat)
		
	default:
		g.logger.Warn("Unhandled message type", zap.String("client_id", client.id), zap.String("type", msg.Type.String()))
	}
}

func (g *WebSocketGateway) handleMovementDelta(client *Client, delta *proto.MovementDelta) {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.entityID == 0 {
		g.logger.Warn("Movement delta from unspawned client", zap.String("client_id", client.id))
		return
	}

	// Update last sequence number for reconciliation
	if delta.Sequence > client.lastSeq {
		client.lastSeq = delta.Sequence
	}

	// Forward movement intent to engine
	g.engine.ProcessMovementIntent(client.entityID, delta)
}

func (g *WebSocketGateway) handleSpawnRequest(client *Client, req *proto.SpawnRequest) {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.entityID != 0 {
		g.logger.Warn("Spawn request from already spawned client", zap.String("client_id", client.id))
		g.sendSpawnResponse(client, false, 0, "client already spawned", 0, 0)
		return
	}

	entityID := g.engine.SpawnEntity(req.EntityType, req.SpawnX, req.SpawnY, client.id)
	if entityID == 0 {
		g.sendSpawnResponse(client, false, 0, "failed to spawn entity", 0, 0)
		return
	}

	client.entityID = entityID
	g.sendSpawnResponse(client, true, entityID, "", req.SpawnX, req.SpawnY)
}

func (g *WebSocketGateway) handleHeartbeat(client *Client, heartbeat *proto.Heartbeat) {
	// TODO: Update last heartbeat time for client
	g.logger.Debug("Received heartbeat", zap.String("client_id", client.id))
}

func (g *WebSocketGateway) sendSpawnResponse(client *Client, success bool, entityID uint32, errorMsg string, x, y float64) {
	response := &proto.Message{
		Type: proto.MessageType_SPAWN_RESPONSE,
		Payload: &proto.Message_SpawnResponse{
			SpawnResponse: &proto.SpawnResponse{
				Success:      success,
				EntityId:     entityID,
				ErrorMessage: errorMsg,
				SpawnX:       x,
				SpawnY:       y,
			},
		},
	}

	data, err := g.codec.Encode(response)
	if err != nil {
		g.logger.Error("Failed to encode spawn response", zap.String("client_id", client.id), zap.Error(err))
		return
	}

	select {
	case client.sendChan <- data:
	default:
		g.logger.Warn("Send buffer full, dropping spawn response", zap.String("client_id", client.id))
	}
}

func (g *WebSocketGateway) BroadcastToClient(clientID string, data []byte) {
	if client, ok := g.clients.Load(clientID); ok {
		if c, ok := client.(*Client); ok {
			select {
			case c.sendChan <- data:
			default:
				g.logger.Warn("Send buffer full, dropping message", zap.String("client_id", clientID))
			}
		}
	}
}

func (g *WebSocketGateway) generateClientID() string {
	return fmt.Sprintf("client_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

func (c *Client) Close() {
	close(c.closeChan)
	if c.conn != nil {
		c.conn.Close()
	}
	close(c.sendChan)
}
