import React from 'react';
import SpatialCanvas from "../canvas/SpatialCanvas";
import TopHUD from "../components/TopHUD";
import MiniMap from "../components/MiniMap";
import Legend from "../components/Legend";
import StatsPanel from "../components/StatsPanel";
import NetworkManager from "../components/NetworkManager";
import styles from './SpatialWorld.module.css';

const WS_URL = 'ws://localhost:8081/ws';

export default function SpatialWorld() {
  return (
    <div className={styles.spatialWorld}>
      {/* Network Manager - handles WebSocket connection and entity synchronization */}
      <NetworkManager wsUrl={WS_URL} />
      
      {/* 3D Canvas - renders the main scene */}
      <SpatialCanvas />
      
      {/* UI Overlay Components */}
      <TopHUD />
      <Legend />
      <StatsPanel />
      <MiniMap />
    </div>
  );
}
