import { create } from 'zustand';
import { subscribeWithSelector } from 'zustand/middleware';

// Entity state structure
const createEntityState = (data) => ({
  id: data.id,
  position: { x: data.position?.x || 0, y: data.position?.y || 0 },
  velocity: { x: data.velocity?.x || 0, y: data.velocity?.y || 0 },
  rotation: data.rotation || 0,
  lastUpdate: data.lastUpdate || Date.now(),
  // Interpolation state
  interpolatedPosition: { x: data.position?.x || 0, y: data.position?.y || 0 },
  previousPosition: { x: data.position?.x || 0, y: data.position?.y || 0 },
  interpolationTime: 0,
  // Visual properties
  type: data.type || 'default',
  color: data.color || '#4a90e2',
  isVisible: true,
  isInAOI: false,
});

const useEntityStore = create(
  subscribeWithSelector((set, get) => ({
    // Core entity storage
    entities: new Map(),
    localPlayerId: null,
    
    // Performance metrics
    entityCount: 0,
    aoiEntityCount: 0,
    
    // Connection state
    isConnected: false,
    latency: 0,
    tickRate: 30, // Default tick rate from server
    
    // Actions
    setLocalPlayerId: (id) => set({ localPlayerId: id }),
    
    setConnectionState: (isConnected, latency = 0) => set({ 
      isConnected, 
      latency 
    }),
    
    setTickRate: (rate) => set({ tickRate: rate }),
    
    // Update entities from server snapshot
    updateFromSnapshot: (snapshot) => {
      const state = get();
      const entities = new Map(state.entities);
      let aoiCount = 0;
      
      // Process full entity states
      if (snapshot.entities) {
        snapshot.entities.forEach(serverEntity => {
          const existingEntity = entities.get(serverEntity.id);
          
          if (existingEntity) {
            // Store current position as previous for interpolation
            existingEntity.previousPosition = { 
              ...existingEntity.interpolatedPosition 
            };
            existingEntity.interpolationTime = 0;
          }
          
          const entity = createEntityState(serverEntity);
          
          // Determine entity type and color based on relationship to local player
          if (state.localPlayerId === serverEntity.id) {
            entity.type = 'local';
            entity.color = '#2ecc71'; // Green for local player
            entity.isInAOI = true;
            aoiCount++;
          } else {
            entity.type = 'remote';
            entity.color = '#e74c3c'; // Red for remote players
            entity.isInAOI = true; // Assume all snapshot entities are in AOI
            aoiCount++;
          }
          
          entities.set(serverEntity.id, entity);
        });
      }
      
      // Process movement deltas (more efficient than full updates)
      if (snapshot.deltas) {
        snapshot.deltas.forEach(delta => {
          const existingEntity = entities.get(delta.id);
          
          if (existingEntity) {
            // Store current position for interpolation
            existingEntity.previousPosition = { 
              ...existingEntity.interpolatedPosition 
            };
            existingEntity.interpolationTime = 0;
            
            // Update server state
            existingEntity.position = { 
              x: delta.position?.x || existingEntity.position.x,
              y: delta.position?.y || existingEntity.position.y
            };
            existingEntity.velocity = { 
              x: delta.velocity?.x || existingEntity.velocity.x,
              y: delta.velocity?.y || existingEntity.velocity.y
            };
            existingEntity.rotation = delta.rotation || existingEntity.rotation;
            existingEntity.lastUpdate = delta.timestamp || Date.now();
          }
        });
      }
      
      set({ 
        entities, 
        entityCount: entities.size,
        aoiEntityCount: aoiCount
      });
    },
    
    // Remove entities that are no longer in the snapshot
    removeEntities: (entityIds) => {
      const state = get();
      const entities = new Map(state.entities);
      
      entityIds.forEach(id => {
        entities.delete(id);
      });
      
      set({ 
        entities, 
        entityCount: entities.size
      });
    },
    
    // Update entity AOI status
    updateAOIStatus: (entityIds, isInAOI) => {
      const state = get();
      const entities = new Map(state.entities);
      let aoiCount = 0;
      
      // Reset all AOI status first
      entities.forEach(entity => {
        entity.isInAOI = false;
      });
      
      // Set AOI status for specified entities
      entityIds.forEach(id => {
        const entity = entities.get(id);
        if (entity) {
          entity.isInAOI = isInAOI;
        }
      });
      
      // Recalculate AOI count
      entities.forEach(entity => {
        if (entity.isInAOI) aoiCount++;
      });
      
      set({ 
        entities, 
        aoiEntityCount: aoiCount
      });
    },
    
    // Get entity by ID
    getEntity: (id) => {
      const state = get();
      return state.entities.get(id);
    },
    
    // Get local player entity
    getLocalPlayer: () => {
      const state = get();
      if (!state.localPlayerId) return null;
      return state.entities.get(state.localPlayerId);
    },
    
    // Get all visible entities
    getVisibleEntities: () => {
      const state = get();
      return Array.from(state.entities.values()).filter(entity => entity.isVisible);
    },
    
    // Get entities in AOI
    getAOIEntities: () => {
      const state = get();
      return Array.from(state.entities.values()).filter(entity => entity.isInAOI);
    },
    
    // Clear all entities (for disconnection)
    clearEntities: () => set({ 
      entities: new Map(), 
      entityCount: 0, 
      aoiEntityCount: 0 
    }),
  }))
);

export default useEntityStore;
