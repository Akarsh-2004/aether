import { useEffect, useRef, useCallback } from 'react';
import useEntityStore from '../state/entityStore';
import WebSocketClient from '../network/WebSocketClient';

export const useWebSocket = (url) => {
  const wsClientRef = useRef(null);
  const {
    updateFromSnapshot,
    setConnectionState,
    clearEntities,
    setLocalPlayerId
  } = useEntityStore();

  // Initialize WebSocket connection
  useEffect(() => {
    if (!url) return;

    const handleConnectionChange = (isConnected) => {
      setConnectionState(isConnected);
      
      if (!isConnected) {
        // Clear entities when disconnected
        clearEntities();
      }
    };

    const handleMessage = (snapshot) => {
      // Process server snapshot
      updateFromSnapshot(snapshot);
      
      // Set local player ID if not already set (assume first entity is local)
      const state = useEntityStore.getState();
      if (!state.localPlayerId && snapshot.entities?.length > 0) {
        setLocalPlayerId(snapshot.entities[0].id);
      }
    };

    wsClientRef.current = new WebSocketClient(
      url,
      handleMessage,
      handleConnectionChange
    );

    wsClientRef.current.connect();

    return () => {
      if (wsClientRef.current) {
        wsClientRef.current.disconnect();
      }
    };
  }, [url, updateFromSnapshot, setConnectionState, clearEntities, setLocalPlayerId]);

  // Send input to server
  const sendInput = useCallback((input) => {
    if (wsClientRef.current) {
      wsClientRef.current.sendInput(input);
    }
  }, []);

  // Get connection state
  const getConnectionState = useCallback(() => {
    if (wsClientRef.current) {
      return wsClientRef.current.getConnectionState();
    }
    return { isConnected: false, latency: 0, reconnectAttempts: 0 };
  }, []);

  return {
    sendInput,
    getConnectionState
  };
};
