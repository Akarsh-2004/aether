import { decodeWorldSnapshot, encodeClientInput, initProtobuf } from './protobuf.js';

class WebSocketClient {
  constructor(url, onMessage, onConnectionChange) {
    this.url = url;
    this.onMessage = onMessage;
    this.onConnectionChange = onConnectionChange;
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
    this.isConnected = false;
    this.lastPingTime = 0;
    this.latency = 0;
    this.heartbeatInterval = null;
  }

  async connect() {
    try {
      // Initialize protobuf before connecting
      const protobufReady = await initProtobuf();
      if (!protobufReady) {
        throw new Error('Failed to initialize protobuf');
      }

      this.ws = new WebSocket(this.url);
      
      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.isConnected = true;
        this.reconnectAttempts = 0;
        this.onConnectionChange?.(true);
        this.startHeartbeat();
      };

      this.ws.onmessage = (event) => {
        // Handle binary protobuf messages
        if (event.data instanceof Blob) {
          event.data.arrayBuffer().then(buffer => {
            try {
              const snapshot = decodeWorldSnapshot(new Uint8Array(buffer));
              this.onMessage?.(snapshot);
            } catch (error) {
              console.error('Failed to parse message:', error);
            }
          });
        } else {
          // Handle text messages (ping/pong, control messages)
          const data = JSON.parse(event.data);
          if (data.type === 'pong') {
            this.latency = Date.now() - this.lastPingTime;
          }
        }
      };

      this.ws.onclose = () => {
        console.log('WebSocket disconnected');
        this.isConnected = false;
        this.onConnectionChange?.(false);
        this.stopHeartbeat();
        this.handleReconnect();
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.isConnected = false;
        this.onConnectionChange?.(false);
      };

    } catch (error) {
      console.error('Failed to connect:', error);
      this.handleReconnect();
    }
  }

  handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
      
      console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
      
      setTimeout(() => {
        this.connect();
      }, delay);
    } else {
      console.error('Max reconnection attempts reached');
    }
  }

  sendInput(input) {
    if (!this.isConnected || !this.ws) {
      console.warn('Cannot send input: WebSocket not connected');
      return;
    }

    try {
      const buffer = encodeClientInput(input);
      this.ws.send(buffer);
    } catch (error) {
      console.error('Failed to send input:', error);
    }
  }

  startHeartbeat() {
    this.heartbeatInterval = setInterval(() => {
      if (this.isConnected && this.ws) {
        this.lastPingTime = Date.now();
        this.ws.send(JSON.stringify({ type: 'ping' }));
      }
    }, 5000); // Ping every 5 seconds
  }

  stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  disconnect() {
    this.stopHeartbeat();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.isConnected = false;
    this.onConnectionChange?.(false);
  }

  getConnectionState() {
    return {
      isConnected: this.isConnected,
      latency: this.latency,
      reconnectAttempts: this.reconnectAttempts
    };
  }
}

export default WebSocketClient;
