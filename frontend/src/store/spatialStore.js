import { create } from "zustand";

export const useSpatialStore = create(set => ({
  city: "Global",
  entities: generateEntities(),

  setCity: city => set({ city }),
  updateEntities: entities => set({ entities }),
}));

function generateEntities() {
  return Array.from({ length: 1500 }, () => ({
    x: Math.random() * window.innerWidth,
    y: Math.random() * window.innerHeight,
    vx: (Math.random() - 0.5) * 0.7,
    vy: (Math.random() - 0.5) * 0.7,
    density: Math.random(),
  }));
}
