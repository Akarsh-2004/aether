import React from 'react';
import { useSpatialStore } from '../store/useSpatialStore';
import { Shield, Activity, Database, Cpu } from 'lucide-react';

export default function CommandHUD() {
    const { session, activeCity, entities, systemLoad } = useSpatialStore();

    return (
        <div className="absolute inset-0 pointer-events-none p-6 flex flex-col justify-between z-10 font-['Plus_Jakarta_Sans']">
            {/* Top Bar */}
            <div className="flex justify-between items-start">
                <div className="bg-white/5 backdrop-blur-xl border border-white/10 p-5 rounded-lg pointer-events-auto">
                    <div className="flex items-center gap-3">
                        <div className="p-2 bg-cyan-500/20 rounded">
                            <Shield className="text-cyan-400" size={20} />
                        </div>
                        <div>
                            <h1 className="text-xl font-extrabold tracking-tight text-white leading-none italic uppercase">
                                Aether<span className="text-cyan-400">Sync</span>
                            </h1>
                            <p className="text-[10px] font-mono text-white/30 tracking-[0.2em] mt-1 uppercase">Spatial Indexing Protocol</p>
                        </div>
                    </div>
                </div>

                <div className="flex gap-3 pointer-events-auto font-mono">
                    <StatusCard label="SESSION_ID" value={session.id} />
                    <StatusCard label="UPLINK" value="STABLE" color="text-emerald-400" />
                </div>
            </div>

            {/* Middle Section: Floating Legend */}
            <div className="absolute left-6 top-1/2 -translate-y-1/2 bg-black/40 border border-white/5 p-4 rounded-md backdrop-blur-sm pointer-events-auto space-y-4">
                <p className="text-[9px] font-bold text-white/20 uppercase tracking-widest">Density Spectrum</p>
                <div className="space-y-2">
                    <LegendItem color="bg-cyan-400" label="Low" />
                    <LegendItem color="bg-indigo-400" label="Nominal" />
                    <LegendItem color="bg-rose-500" label="Congested" />
                </div>
            </div>

            {/* Bottom Bar */}
            <div className="flex justify-between items-end">
                <div className="bg-black/60 border border-white/10 p-4 rounded-xl backdrop-blur-2xl pointer-events-auto min-w-[300px]">
                    <div className="flex justify-between text-[10px] text-white/40 font-bold mb-3 uppercase tracking-tighter">
                        <span>Node Analytics: {activeCity}</span>
                        <span>Load: {systemLoad}%</span>
                    </div>
                    <div className="w-full h-1 bg-white/5 rounded-full overflow-hidden">
                        <div className="h-full bg-cyan-500 transition-all duration-1000" style={{width: `${systemLoad}%`}} />
                    </div>
                    <div className="mt-4 flex justify-between">
                        <Metric icon={<Activity size={12}/>} value={`${entities.length} pts`} />
                        <Metric icon={<Database size={12}/>} value="4.2 GB/s" />
                        <Metric icon={<Cpu size={12}/>} value="0.4ms" />
                    </div>
                </div>

                {/* City Selector */}
                <div className="flex gap-2 pointer-events-auto">
                    {['Global', 'Mumbai', 'London', 'Tokyo'].map(city => (
                        <button 
                            key={city}
                            className="px-4 py-2 bg-white/5 hover:bg-white/10 border border-white/10 rounded-md text-[11px] font-bold uppercase tracking-wider transition-all hover:text-cyan-400"
                        >
                            {city}
                        </button>
                    ))}
                </div>
            </div>
        </div>
    );
}

const StatusCard = ({ label, value, color = "text-white" }) => (
    <div className="bg-white/5 border border-white/10 px-4 py-2 rounded flex flex-col items-end backdrop-blur-lg">
        <span className="text-[8px] text-white/30 font-bold">{label}</span>
        <span className={`text-xs font-bold ${color}`}>{value}</span>
    </div>
);

const LegendItem = ({ color, label }) => (
    <div className="flex items-center gap-2">
        <div className={`w-1.5 h-1.5 rounded-full ${color}`} />
        <span className="text-[10px] text-white/60 font-medium">{label}</span>
    </div>
);

const Metric = ({ icon, value }) => (
    <div className="flex items-center gap-1.5 text-white/60 font-mono text-[10px]">
        {icon} <span>{value}</span>
    </div>
);