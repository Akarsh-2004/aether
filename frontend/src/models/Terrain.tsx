import React, { useMemo, Suspense } from 'react';
import { useGLTF } from '@react-three/drei';
import * as THREE from 'three';

// Cache for the terrain model to avoid re-loading
let terrainModelCache: THREE.Group | null = null;

interface TerrainProps {
  position?: [number, number, number];
  scale?: number;
  receiveShadows?: boolean;
}

// Preloader component for the terrain model
const TerrainModelLoader = () => {
  // Load the FBX terrain model
  const gltf = useGLTF('/models/terrain.fbx');
  
  if (!gltf) return null;
  
  // Cache the model for future use
  if (!terrainModelCache) {
    terrainModelCache = gltf.scene;
  }
  
  return null;
};

// Main Terrain component
export const Terrain = React.memo<TerrainProps>(({
  position = [0, 0, 0],
  scale = 1,
  receiveShadows = true
}) => {
  // Use cached model if available, otherwise load it
  const model = useMemo(() => {
    if (terrainModelCache) {
      const clonedModel = terrainModelCache.clone();
      
      // Apply neutral material for architectural feel
      clonedModel.traverse((child) => {
        if (child instanceof THREE.Mesh) {
          child.material = new THREE.MeshStandardMaterial({
            color: 0xf5f5f5, // Light gray
            metalness: 0.0,
            roughness: 0.9,
          });
          
          // Enable shadow receiving
          child.receiveShadow = receiveShadows;
          child.castShadow = false; // Terrain doesn't cast shadows on itself
        }
      });
      
      return clonedModel;
    }
    return null;
  }, [receiveShadows]);

  if (!model) {
    // Fallback to a simple plane while model loads
    return (
      <mesh position={position} rotation={[-Math.PI / 2, 0, 0]} scale={[scale, scale, scale]}>
        <planeGeometry args={[100, 100]} />
        <meshStandardMaterial 
          color="#f5f5f5" 
          roughness={0.9}
        />
      </mesh>
    );
  }

  return (
    <primitive 
      object={model} 
      position={position}
      scale={[scale, scale, scale]}
    />
  );
});

// Component to preload the terrain model
export const TerrainPreloader = () => {
  return (
    <Suspense fallback={null}>
      <TerrainModelLoader />
    </Suspense>
  );
};

Terrain.displayName = 'Terrain';
