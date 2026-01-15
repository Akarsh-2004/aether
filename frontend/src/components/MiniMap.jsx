import React from 'react';
import { useSpatialStore } from '../store/useSpatialStore';

const CITIES = ['Global Mesh', 'Mumbai', 'London', 'Tokyo', 'San Francisco'];

export default function MiniMap() {
    const { currentCity, setCity } = useSpatialStore();

    return (
        <div className="absolute bottom-10 left-1/2 -translate-x-1/2 bg-white/5 backdrop-blur-xl border border-white/10 p-2 rounded-2xl flex gap-1 pointer-events-auto z-20 shadow-2xl">
            {CITIES.map((city) => (
                <button
                    key={city}
                    onClick={() => setCity(city)}
                    className={`
                        px-6 py-3 rounded-xl text-[10px] font-bold uppercase tracking-widest transition-all duration-300
                        ${currentCity === city 
                            ? 'bg-cyan-500 text-black shadow-[0_0_20px_rgba(34,211,238,0.3)] scale-105' 
                            : 'text-white/40 hover:bg-white/5 hover:text-white'}
                    `}
                >
                    {city}
                </button>
            ))}
        </div>
    );
}
