import React, { useEffect, useRef } from 'react';
import { useSpatialStore } from '../store/useSpatialStore';
import { useWebSocket } from '../hooks/useWebSocket';

const NetworkManager = ({ wsUrl }) => {
  const { updateFromSnapshot, setConnectionState, clearEntities, setLocalPlayerId } = useSpatialStore();
  const frameRef = useRef();
  const lastUpdateRef = useRef(0);

  // Handle incoming server snapshots
  const handleMessage = (snapshot) => {
    // Update entity store with server data
    updateFromSnapshot(snapshot);
    
    // Track update timing for interpolation
    lastUpdateRef.current = performance.now();
  };

  // Handle connection state changes
  const handleConnectionChange = (isConnected) => {
    setConnectionState(isConnected);
    
    if (!isConnected) {
      // Clear entities when disconnected to prevent stale data
      clearEntities();
    }
  };

  // Initialize WebSocket connection
  const { sendInput, getConnectionState } = useWebSocket(
    wsUrl,
    handleMessage,
    handleConnectionChange
  );

  // Client-side prediction and interpolation loop
  useEffect(() => {
    const tick = () => {
      const now = performance.now();
      const state = useSpatialStore.getState();
      
      // Only interpolate if we have entities and connection
      if (state.entities.length > 0 && state.isConnected) {
        const timeSinceLastUpdate = now - lastUpdateRef.current;
        
        // Simple linear interpolation for smooth movement
        // In a real implementation, this would be more sophisticated
        const interpolatedEntities = state.entities.map(entity => {
          // Apply velocity-based prediction
          const predictionTime = Math.min(timeSinceLastUpdate / 1000, 0.1); // Cap at 100ms
          
          return {
            ...entity,
            predictedX: entity.x + (entity.vx || 0) * predictionTime,
            predictedY: entity.y + (entity.vy || 0) * predictionTime
          };
        });
        
        // Update store with interpolated positions
        useSpatialStore.getState().updateEntities(interpolatedEntities);
      }
      
      frameRef.current = requestAnimationFrame(tick);
    };
    
    frameRef.current = requestAnimationFrame(tick);
    
    return () => {
      if (frameRef.current) {
        cancelAnimationFrame(frameRef.current);
      }
    };
  }, []);

  // Send periodic heartbeat to maintain connection
  useEffect(() => {
    const heartbeat = setInterval(() => {
      const connectionState = getConnectionState();
      if (connectionState.isConnected) {
        sendInput({ type: 'heartbeat', timestamp: Date.now() });
      }
    }, 5000); // Send heartbeat every 5 seconds

    return () => clearInterval(heartbeat);
  }, [sendInput, getConnectionState]);

  return null; // This component doesn't render anything
};

export default NetworkManager;
