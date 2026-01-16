import React from 'react';
import { HemisphereLight, DirectionalLight, AmbientLight } from 'three';
import { useThree } from '@react-three/fiber';

/**
 * Professional lighting setup for architectural simulation
 * 
 * Design rationale:
 * - Hemisphere light provides natural sky/ground ambient lighting
 * - Directional light creates soft, realistic shadows
 * - Ambient light fills shadows for clean, architectural feel
 * - Intensity tuned for white base theme without harsh contrast
 */
export const Lighting = () => {
  const { scene } = useThree();

  // Add lights to scene when component mounts
  React.useEffect(() => {
    // Hemisphere light simulates natural outdoor lighting
    const hemisphereLight = new HemisphereLight('#e8f4f8', '#f0f0f0', 0.4);
    hemisphereLight.position.set(0, 50, 0);
    scene.add(hemisphereLight);

    // Directional light simulates sun with soft shadows
    const directionalLight = new DirectionalLight(0xffffff, 0.8);
    directionalLight.position.set(20, 30, 10);
    directionalLight.castShadow = true;
    directionalLight.shadow.mapSize.width = 2048;
    directionalLight.shadow.mapSize.height = 2048;
    directionalLight.shadow.camera.far = 100;
    directionalLight.shadow.camera.left = -50;
    directionalLight.shadow.camera.right = 50;
    directionalLight.shadow.camera.top = 50;
    directionalLight.shadow.camera.bottom = -50;
    directionalLight.shadow.bias = -0.0001;
    scene.add(directionalLight);

    // Soft ambient light to fill shadows
    const ambientLight = new AmbientLight('#fafafa', 0.3);
    scene.add(ambientLight);

    return () => {
      scene.remove(hemisphereLight);
      scene.remove(directionalLight);
      scene.remove(ambientLight);
    };
  }, [scene]);

  return null;
};
