import { useEffect, useRef } from "react";
import { useSpatialStore } from "../store/spatialStore";

export default function SpatialCanvas() {
  const canvasRef = useRef();
  const { entities } = useSpatialStore();

  useEffect(() => {
    const canvas = canvasRef.current;
    const ctx = canvas.getContext("2d");

    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;

    function loop() {
      ctx.clearRect(0, 0, canvas.width, canvas.height);

      entities.forEach(e => {
        e.x += e.vx;
        e.y += e.vy;

        if (e.x < 0 || e.x > canvas.width) e.vx *= -1;
        if (e.y < 0 || e.y > canvas.height) e.vy *= -1;

        const color =
          e.density < 0.33 ? "#4CC9F0" :
          e.density < 0.66 ? "#F9C74F" :
          "#F94144";

        ctx.beginPath();
        ctx.fillStyle = color;
        ctx.arc(e.x, e.y, 2, 0, Math.PI * 2);
        ctx.fill();
      });

      requestAnimationFrame(loop);
    }

    loop();
  }, [entities]);

  return <canvas ref={canvasRef} className="absolute inset-0" />;
}
