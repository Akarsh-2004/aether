import React, { useRef, useMemo, Suspense } from 'react';
import { useFrame } from '@react-three/fiber';
import { useGLTF, useOBJ } from '@react-three/drei';
import * as THREE from 'three';

// Cache for the loaded human model to avoid re-loading
let humanModelCache: THREE.Group | null = null;

interface HumanBodyProps {
  position?: [number, number, number];
  rotation?: [number, number, number];
  scale?: number;
  color?: string;
  entityId?: string;
  isLocalPlayer?: boolean;
}

// Preloader component for the human model
const HumanModelLoader = () => {
  // Load the OBJ model
  const obj = useOBJ('/models/FinalBaseMesh.obj');
  
  if (!obj) return null;
  
  // Cache the model for future use
  if (!humanModelCache) {
    humanModelCache = obj.scene;
  }
  
  return null;
};

// Main HumanBody component
export const HumanBody = React.memo<HumanBodyProps>(({
  position = [0, 0, 0],
  rotation = [0, 0, 0],
  scale = 1,
  color = '#f0f0f0',
  entityId,
  isLocalPlayer = false
}) => {
  const meshRef = useRef<THREE.Group>(null);
  
  // Use cached model if available, otherwise load it
  const model = useMemo(() => {
    if (humanModelCache) {
      return humanModelCache.clone();
    }
    return null;
  }, []);

  // Smooth interpolation for position updates
  useFrame((state, delta) => {
    if (meshRef.current && model) {
      // Smooth movement interpolation (15% lerp per frame)
      meshRef.current.position.x += (position[0] - meshRef.current.position.x) * 0.15;
      meshRef.current.position.y += (position[1] - meshRef.current.position.y) * 0.15;
      meshRef.current.position.z += (position[2] - meshRef.current.position.z) * 0.15;
    }
  });

  // Create material with the specified color
  const material = useMemo(() => {
    return new THREE.MeshStandardMaterial({
      color: new THREE.Color(color),
      metalness: 0.1,
      roughness: 0.8,
    });
  }, [color]);

  if (!model) {
    // Fallback to a simple box while model loads
    return (
      <mesh ref={meshRef} position={position} rotation={rotation} scale={[scale, scale, scale]}>
        <boxGeometry args={[0.5, 1.8, 0.3]} />
        <meshStandardMaterial color={color} />
      </mesh>
    );
  }

  return (
    <group 
      ref={meshRef} 
      position={position} 
      rotation={rotation} 
      scale={[scale, scale, scale]}
    >
      <primitive 
        object={model} 
        material={material}
      />
      {/* Add a subtle indicator for local player */}
      {isLocalPlayer && (
        <mesh position={[0, 2.2, 0]}>
          <sphereGeometry args={[0.1, 8, 8]} />
          <meshStandardMaterial 
            color="#4a90e2" 
            emissive="#4a90e2" 
            emissiveIntensity={0.3}
          />
        </mesh>
      )}
    </group>
  );
});

// Component to preload the human model
export const HumanBodyPreloader = () => {
  return (
    <Suspense fallback={null}>
      <HumanModelLoader />
    </Suspense>
  );
};

HumanBody.displayName = 'HumanBody';
