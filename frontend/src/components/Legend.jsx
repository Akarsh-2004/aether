export default function Legend() {
  return (
    <div className="absolute top-16 left-4 bg-black/50 px-3 py-2 rounded text-xs">
      <div><span className="text-blue-400">●</span> Low Traffic</div>
      <div><span className="text-yellow-400">●</span> Medium</div>
      <div><span className="text-red-400">●</span> Congested</div>
    </div>
  );
}
