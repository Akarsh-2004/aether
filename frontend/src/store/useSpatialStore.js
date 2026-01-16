import { create } from 'zustand';
import { useWebSocket } from '../hooks/useWebSocket';

// Generate initial entities for demo purposes
const generateInitialEntities = (count) => {
    return Array.from({ length: count }, (_, i) => ({
        id: `entity-${i}`,
        x: (Math.random() - 0.5) * 100, // World coordinates (-50 to 50)
        y: (Math.random() - 0.5) * 100,
        vx: (Math.random() - 0.5) * 0.5,
        vy: (Math.random() - 0.5) * 0.5,
        density: Math.random(),
        size: Math.random() * 0.5 + 0.5,
        type: 'vehicle' // Could be 'vehicle', 'node', etc.
    }));
};

export const useSpatialStore = create((set, get) => ({
    // World state
    city: 'Global Mesh',
    entities: generateInitialEntities(100), // Start with fewer entities for performance
    localPlayerId: null,
    
    // Connection state
    isConnected: false,
    latency: 0,
    reconnectAttempts: 0,
    
    // Stats
    stats: {
        latency: '0ms',
        load: '0%',
        activeNodes: 0,
        tickRate: '20Hz'
    },

    // Actions
    setCity: (cityName) => set({ 
        city: cityName,
        entities: generateInitialEntities(50 + Math.floor(Math.random() * 100))
    }),

    updateEntities: (newEntities) => set({ entities: newEntities }),

    // Update from server snapshot
    updateFromSnapshot: (snapshot) => {
        const state = get();
        const entities = snapshot.entities || [];
        
        set({
            entities: entities,
            stats: {
                ...state.stats,
                activeNodes: entities.length,
                latency: `${snapshot.latency || 0}ms`,
                tickRate: `${snapshot.tickRate || 20}Hz`
            }
        });
    },

    // Set connection state
    setConnectionState: (isConnected) => set({ isConnected }),

    // Set local player ID
    setLocalPlayerId: (playerId) => set({ localPlayerId: playerId }),

    // Clear all entities
    clearEntities: () => set({ entities: [] }),

    // Update stats
    updateStats: (newStats) => set(state => ({
        stats: { ...state.stats, ...newStats }
    }))
}));

