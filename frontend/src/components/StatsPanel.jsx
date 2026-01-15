import { useSpatialStore } from "../store/spatialStore";

export default function StatsPanel() {
  const count = useSpatialStore(s => s.entities.length);

  return (
    <div className="absolute top-16 right-4 bg-black/50 px-3 py-2 rounded text-xs">
      Active Entities: {count}
    </div>
  );
}
