import { useSpatialStore } from "../store/spatialStore";

const cities = ["Global", "Mumbai", "Delhi", "London", "New York", "Tokyo"];

export default function MiniMap() {
  const setCity = useSpatialStore(s => s.setCity);

  return (
    <div className="absolute bottom-4 left-1/2 -translate-x-1/2
      bg-black/60 px-4 py-2 rounded-lg flex gap-4 z-10">
      {cities.map(c => (
        <button
          key={c}
          onClick={() => setCity(c)}
          className="text-xs text-gray-300 hover:text-white"
        >
          {c}
        </button>
      ))}
    </div>
  );
}
