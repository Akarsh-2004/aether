import React, { useEffect } from 'react';
import { useSpatialStore } from '../store/useSpatialStore';
import { Activity, Shield, Zap, Cpu, Globe, ArrowUpRight } from 'lucide-react';

export default function CommandHUD() {
    const { telemetry, currentCity, tickMetrics, nodeId } = useSpatialStore();

    useEffect(() => {
        const timer = setInterval(tickMetrics, 1500);
        return () => clearInterval(timer);
    }, []);

    return (
        <div className="absolute inset-0 pointer-events-none p-10 flex flex-col justify-between z-10 font-sans">
            {/* TOP NAVIGATION HUD */}
            <div className="flex justify-between items-start">
                <div className="bg-white/5 backdrop-blur-2xl border border-white/10 p-6 rounded-2xl pointer-events-auto shadow-2xl transition-all hover:bg-white/10">
                    <div className="flex items-center gap-5">
                        <div className="p-3 bg-cyan-500/20 rounded-xl border border-cyan-500/30">
                            <Shield className="text-cyan-400" size={26} />
                        </div>
                        <div>
                            <h1 className="text-3xl font-black italic tracking-tighter text-white uppercase leading-none">
                                AETHER<span className="text-cyan-400">SYNC</span>
                            </h1>
                            <p className="text-[10px] font-mono text-white/40 tracking-[0.3em] uppercase mt-2 font-bold">
                                Spatial Synchronization Cluster • {nodeId}
                            </p>
                        </div>
                    </div>
                </div>

                <div className="flex gap-4 pointer-events-auto">
                    <TelemetryBox label="Latency" value={telemetry.latency} icon={<Zap size={14}/>} color="text-cyan-400" />
                    <TelemetryBox label="Active Nodes" value={telemetry.activeNodes} icon={<Globe size={14}/>} />
                </div>
            </div>

            {/* BOTTOM ANALYTICS HUD */}
            <div className="flex justify-between items-end">
                <div className="bg-black/60 border border-white/10 p-6 rounded-3xl backdrop-blur-3xl pointer-events-auto min-w-[350px]">
                    <div className="flex justify-between items-center mb-6 px-1">
                        <span className="text-[11px] font-bold text-white/40 uppercase tracking-widest italic">Regional Index: {currentCity}</span>
                        <div className="flex items-center gap-2">
                            <div className="w-2 h-2 bg-emerald-500 rounded-full animate-pulse shadow-[0_0_8px_#10b981]" />
                            <span className="text-[10px] font-mono text-emerald-500 font-bold uppercase">Uplink Stable</span>
                        </div>
                    </div>
                    
                    <div className="space-y-4">
                        <div className="flex justify-between text-[11px] font-mono text-white/30 px-1">
                            <div className="flex items-center gap-2"><Cpu size={12}/> Load: {telemetry.cpuUsage}</div>
                            <div className="flex items-center gap-2"><ArrowUpRight size={12}/> {telemetry.throughput}</div>
                        </div>
                        <div className="w-full h-1 bg-white/5 rounded-full overflow-hidden">
                            <div 
                                className="h-full bg-gradient-to-r from-cyan-500 to-indigo-500 transition-all duration-1000" 
                                style={{ width: telemetry.cpuUsage }} 
                            />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

function TelemetryBox({ label, value, icon, color = "text-white" }) {
    return (
        <div className="bg-white/5 backdrop-blur-2xl border border-white/10 px-7 py-5 rounded-2xl flex flex-col items-end shadow-xl">
            <div className="flex items-center gap-2 text-white/20 mb-1">
                {icon} <span className="text-[9px] font-bold uppercase tracking-[0.2em]">{label}</span>
            </div>
            <div className={`text-2xl font-mono font-bold ${color} tracking-tighter`}>{value}</div>
        </div>
    );
}
