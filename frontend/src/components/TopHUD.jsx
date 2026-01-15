import { useSpatialStore } from "../store/spatialStore";

export default function TopHUD() {
  const city = useSpatialStore(s => s.city);

  return (
    <div className="absolute top-4 left-1/2 -translate-x-1/2
      bg-black/60 backdrop-blur px-6 py-2 rounded-lg text-sm z-10">
      AetherSync • City: {city} • Real-Time Spatial Engine
    </div>
  );
}
