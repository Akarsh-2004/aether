import React, { useRef, useEffect } from 'react';
import { useSpatialStore } from '../store/useSpatialStore';

const CoreEngine = () => {
    const canvasRef = useRef(null);
    const entities = useSpatialStore(state => state.entities);

    useEffect(() => {
        const canvas = canvasRef.current;
        const ctx = canvas.getContext('2d');
        let frame;

        const resize = () => {
            canvas.width = window.innerWidth;
            canvas.height = window.innerHeight;
        };
        window.addEventListener('resize', resize);
        resize();

        const render = () => {
            ctx.fillStyle = 'rgba(5, 5, 8, 0.2)';
            ctx.fillRect(0, 0, canvas.width, canvas.height);

            entities.forEach(e => {
                e.x += e.vx; e.y += e.vy;
                if (e.x < 0 || e.x > canvas.width) e.vx *= -1;
                if (e.y < 0 || e.y > canvas.height) e.vy *= -1;

                let color = '#4cc9f0'; 
                if (e.density > 0.8) color = '#f72585'; 
                else if (e.density > 0.5) color = '#fee440';

                ctx.beginPath();
                ctx.arc(e.x, e.y, e.size, 0, Math.PI * 2);
                ctx.fillStyle = color;
                
                if (e.density > 0.8) {
                    ctx.shadowBlur = 10;
                    ctx.shadowColor = color;
                } else {
                    ctx.shadowBlur = 0;
                }
                
                ctx.fill();
            });

            frame = requestAnimationFrame(render);
        };

        render();
        return () => {
            cancelAnimationFrame(frame);
            window.removeEventListener('resize', resize);
        };
    }, [entities]);

    return <canvas ref={canvasRef} className="absolute inset-0 z-0 bg-transparent" />;
};

export default CoreEngine;

