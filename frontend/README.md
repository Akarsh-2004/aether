# Aether Frontend - Realtime Spatial Visualization

A high-performance, engineering-grade frontend for visualizing realtime spatial systems with WebSocket-based multiplayer support.

## Architecture

### Core Components

- **Network Layer**: WebSocket client with Protobuf serialization
- **State Management**: Zustand-based entity store with interpolation
- **Rendering**: Three.js with React Three Fiber for 3D visualization
- **UI**: Minimal HUD with performance metrics

### Performance Features

- **Instanced Rendering**: Efficient rendering for 300+ entities
- **Client-side Interpolation**: Smooth movement between server updates
- **Optimized Camera**: Perspective camera with smooth following
- **Minimal Re-renders**: React hooks optimized for performance

## Tech Stack

- **React 18** with Vite
- **Three.js** + React Three Fiber + Drei
- **Zustand** for state management
- **Protobuf.js** for binary serialization
- **Custom CSS** (no Tailwind for performance)

## Key Features

### Realtime Networking
- Binary WebSocket communication
- Automatic reconnection with exponential backoff
- Heartbeat/ping monitoring
- Protobuf message encoding/decoding

### Entity Management
- Authoritative server model
- Area of Interest (AOI) system
- Entity interpolation and smoothing
- Local player distinction

### 3D Visualization
- Orthographic grid world (1000x1000 units)
- Color-coded entities (green=local, red=remote)
- AOI radius visualization
- Performance-optimized rendering

### HUD Interface
- Connection status and latency
- Entity count metrics
- Tick rate display
- FPS counter
- Expandable control panel

## Development

### Setup
```bash
npm install
npm run dev
```

### Configuration
- WebSocket URL: `WS_URL` in `SpatialCanvas.jsx`
- World size: `WORLD_SIZE` in `WorldScene.jsx`
- AOI radius: Configurable via HUD

### Architecture Decisions

**Camera Choice**: Perspective camera over orthographic for better depth perception and spatial awareness in multiplayer scenarios.

**Rendering Strategy**: Hybrid approach - individual meshes for <50 entities, instanced rendering for larger crowds.

**State Management**: Zustand chosen over Redux for minimal boilerplate and better performance with frequent updates.

**Styling**: Custom CSS instead of Tailwind to reduce bundle size and eliminate runtime CSS processing.

## Performance Considerations

- Canvas resolution reduced by 50% for better FPS
- Antialiasing disabled for performance
- Instanced rendering for entity crowds
- Efficient interpolation algorithms
- Minimal React re-renders using refs and selectors

## Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

Requires WebGL2 support for optimal performance.
