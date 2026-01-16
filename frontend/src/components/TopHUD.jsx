import React from 'react';
import { useSpatialStore } from '../store/useSpatialStore';
import styles from './TopHUD.module.css';

const TopHUD = () => {
  const { city, isConnected, stats, localPlayerId } = useSpatialStore();

  return (
    <div className={styles.topHUD}>
      <div className={styles.header}>
        <h1 className={styles.title}>AetherSync</h1>
        <div className={styles.cityName}>{city}</div>
      </div>
      <div className={styles.statusBar}>
        <div className={`${styles.status} ${isConnected ? styles.connected : styles.disconnected}`}>
          {isConnected ? 'Connected' : 'Disconnected'}
        </div>
        <div className={styles.stats}>
          <span>Latency: {stats.latency}</span>
          <span>Entities: {stats.activeNodes}</span>
          <span>Tick: {stats.tickRate}</span>
          {localPlayerId && <span>ID: {localPlayerId.slice(-8)}</span>}
        </div>
      </div>
    </div>
  );
};

export default TopHUD;
