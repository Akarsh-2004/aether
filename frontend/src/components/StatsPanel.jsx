import React from 'react';
import { useSpatialStore } from '../store/useSpatialStore';
import styles from './StatsPanel.module.css';

const StatsPanel = () => {
  const { stats, isConnected, entities, localPlayerId } = useSpatialStore();

  return (
    <div className={styles.statsPanel}>
      <h3 className={styles.title}>System Statistics</h3>
      <div className={styles.statsGrid}>
        <div className={styles.statItem}>
          <span className={styles.statLabel}>Connection</span>
          <span className={`${styles.statValue} ${isConnected ? styles.connected : styles.disconnected}`}>
            {isConnected ? 'Online' : 'Offline'}
          </span>
        </div>
        <div className={styles.statItem}>
          <span className={styles.statLabel}>Entities</span>
          <span className={styles.statValue}>{entities.length}</span>
        </div>
        <div className={styles.statItem}>
          <span className={styles.statLabel}>Latency</span>
          <span className={styles.statValue}>{stats.latency}</span>
        </div>
        <div className={styles.statItem}>
          <span className={styles.statLabel}>Tick Rate</span>
          <span className={styles.statValue}>{stats.tickRate}</span>
        </div>
        <div className={styles.statItem}>
          <span className={styles.statLabel}>Load</span>
          <span className={styles.statValue}>{stats.load}</span>
        </div>
        {localPlayerId && (
          <div className={styles.statItem}>
            <span className={styles.statLabel}>Player ID</span>
            <span className={styles.statValue}>{localPlayerId.slice(-8)}</span>
          </div>
        )}
      </div>
    </div>
  );
};

export default StatsPanel;
