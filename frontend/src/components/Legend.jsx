import React from 'react';

export default function Legend() {
    return (
        <div className="absolute top-44 left-10 bg-black/40 backdrop-blur-md border border-white/5 p-5 rounded-2xl pointer-events-auto z-10 flex flex-col gap-4">
            <p className="text-[9px] font-black text-white/20 uppercase tracking-[0.3em] mb-1">Traffic Spectrum</p>
            <div className="space-y-4">
                <LegendItem color="bg-[#22d3ee]" label="Optimal" desc="Free Flow" />
                <LegendItem color="bg-[#fbbf24]" label="Warning" desc="Congested" />
                <LegendItem color="bg-[#f43f5e]" label="Critical" desc="Saturation" />
            </div>
            <div className="mt-2 border-t border-white/5 pt-4">
                <div className="flex items-center gap-3">
                    <div className="w-1.5 h-1.5 rounded-full bg-white animate-ping" />
                    <span className="text-[10px] font-mono text-white/40 italic">Scanning Spatial Nodes...</span>
                </div>
            </div>
        </div>
    );
}

function LegendItem({ color, label, desc }) {
    return (
        <div className="flex items-center gap-4 group cursor-default">
            <div className={`w-2.5 h-2.5 rounded-full ${color} shadow-lg transition-transform group-hover:scale-125`} />
            <div>
                <div className="text-[10px] font-bold text-white uppercase leading-none">{label}</div>
                <div className="text-[8px] text-white/30 uppercase mt-0.5">{desc}</div>
            </div>
        </div>
    );
}
