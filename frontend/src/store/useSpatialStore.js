import { create } from 'zustand';
const generateInitialTraffic = (count) => {
    return Array.from({ length: count }, (_, i) => ({
        id: `node-${i}`,
        x: Math.random() * window.innerWidth,
        y: Math.random() * window.innerHeight,
        vx: (Math.random() - 0.5) * 1.5,
        vy: (Math.random() - 0.5) * 1.5,
        density: Math.random(), 
        size: Math.random() * 2 + 1
    }));
};

export const useSpatialStore = create((set) => ({
    city: 'Global Mesh',
    entities: generateInitialTraffic(2500), 
    stats: {
        latency: '12ms',
        load: '34%',
        activeNodes: 2500
    },
    
    setCity: (cityName) => set({ 
        city: cityName, 
        entities: generateInitialTraffic(1500 + Math.floor(Math.random() * 1500)) 
    }),
    updateEntities: (newBatch) => set({ entities: newBatch })
}));

