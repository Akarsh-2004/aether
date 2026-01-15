import SpatialCanvas from "../canvas/SpatialCanvas";
import TopHUD from "../components/TopHUD";
import MiniMap from "../components/MiniMap";
import Legend from "../components/Legend";
import StatsPanel from "../components/StatsPanel";

export default function SpatialWorld() {
  return (
    <div className="relative w-full h-full">
      <TopHUD />
      <SpatialCanvas />
      <Legend />
      <StatsPanel />
      <MiniMap />
    </div>
  );
}
