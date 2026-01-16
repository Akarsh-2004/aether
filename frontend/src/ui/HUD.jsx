import React, { useState, useEffect } from 'react';
import useEntityStore from '../state/entityStore';
import './HUD.css';

const HUD = () => {
  const [showAOI, setShowAOI] = useState(false);
  const [expanded, setExpanded] = useState(false);
  
  const {
    isConnected,
    latency,
    entityCount,
    aoiEntityCount,
    tickRate,
    localPlayerId,
    getLocalPlayer
  } = useEntityStore();

  const localPlayer = useEntityStore(getLocalPlayer);

  // Format latency display
  const formatLatency = (ms) => {
    if (ms < 50) return `${ms}ms`;
    if (ms < 100) return `${ms}ms`;
    return `${ms}ms`;
  };

  // Get connection status color
  const getConnectionColor = () => {
    if (!isConnected) return '#e74c3c'; // Red
    if (latency < 50) return '#2ecc71'; // Green
    if (latency < 100) return '#f39c12'; // Orange
    return '#e74c3c'; // Red
  };

  // Get connection status text
  const getConnectionText = () => {
    if (!isConnected) return 'Disconnected';
    if (latency < 50) return 'Excellent';
    if (latency < 100) return 'Good';
    return 'Poor';
  };

  return (
    <div className="hud-container">
      {/* Main HUD Bar */}
      <div className="hud-main">
        {/* Connection Status */}
        <div className="hud-section">
          <div 
            className="status-indicator"
            style={{ backgroundColor: getConnectionColor() }}
          />
          <span className="status-text">{getConnectionText()}</span>
          <span className="latency-text">{formatLatency(latency)}</span>
        </div>

        {/* Entity Count */}
        <div className="hud-section">
          <span className="metric-label">Entities:</span>
          <span className="metric-value">{entityCount}</span>
          <span className="metric-sub">({aoiEntityCount} AOI)</span>
        </div>

        {/* Tick Rate */}
        <div className="hud-section">
          <span className="metric-label">Tick:</span>
          <span className="metric-value">{tickRate}Hz</span>
        </div>

        {/* Expand/Collapse Button */}
        <button 
          className="hud-toggle"
          onClick={() => setExpanded(!expanded)}
        >
          {expanded ? '▼' : '▲'}
        </button>
      </div>

      {/* Expanded HUD Panel */}
      {expanded && (
        <div className="hud-expanded">
          {/* Player Info */}
          {localPlayer && (
            <div className="hud-section">
              <h4>Local Player</h4>
              <div className="player-info">
                <span>ID: {localPlayerId}</span>
                <span>Pos: ({Math.round(localPlayer.position.x)}, {Math.round(localPlayer.position.y)})</span>
                <span>Vel: ({localPlayer.velocity.x.toFixed(1)}, {localPlayer.velocity.y.toFixed(1)})</span>
              </div>
            </div>
          )}

          {/* Controls */}
          <div className="hud-section">
            <h4>Controls</h4>
            <div className="controls">
              <label className="control-item">
                <input
                  type="checkbox"
                  checked={showAOI}
                  onChange={(e) => setShowAOI(e.target.checked)}
                />
                Show AOI Radius
              </label>
            </div>
          </div>

          {/* Performance Metrics */}
          <div className="hud-section">
            <h4>Performance</h4>
            <div className="performance-metrics">
              <div className="metric">
                <span>FPS:</span>
                <span id="fps-counter">60</span>
              </div>
              <div className="metric">
                <span>Draw Calls:</span>
                <span id="draw-calls">--</span>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Connection Warning */}
      {!isConnected && (
        <div className="connection-warning">
          <span>⚠️ Connection Lost</span>
          <span>Attempting to reconnect...</span>
        </div>
      )}

      {/* Instructions */}
      <div className="instructions">
        <div>WASD/Arrow Keys: Move</div>
        <div>Mouse: Camera Control</div>
        <div>Scroll: Zoom</div>
      </div>
    </div>
  );
};

export default HUD;
