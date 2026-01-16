import { useRef, useMemo } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { OrbitControls, Grid } from '@react-three/drei';
import * as THREE from 'three';

// World bounds and configuration
const WORLD_SIZE = 1000; // 1000x1000 units
const GRID_SIZE = 100; // Grid divisions
const CAMERA_FOV = 75;
const CAMERA_NEAR = 0.1;
const CAMERA_FAR = 2000;

// Camera controller component
function CameraController({ localPlayerPosition }) {
  const cameraRef = useRef();
  
  useFrame(() => {
    if (cameraRef.current && localPlayerPosition) {
      // Smooth camera follow with slight offset for better view
      const targetX = localPlayerPosition.x;
      const targetY = localPlayerPosition.y;
      const targetZ = 50; // Fixed height for top-down view
      
      // Smooth interpolation
      cameraRef.current.position.x += (targetX - cameraRef.current.position.x) * 0.1;
      cameraRef.current.position.y += (targetY - cameraRef.current.position.y) * 0.1;
      cameraRef.current.position.z += (targetZ - cameraRef.current.position.z) * 0.1;
      
      // Look at the player position
      cameraRef.current.lookAt(targetX, targetY, 0);
    }
  });

  return (
    <perspectiveCamera
      ref={cameraRef}
      fov={CAMERA_FOV}
      near={CAMERA_NEAR}
      far={CAMERA_FAR}
      position={[0, 0, 50]}
    />
  );
}

// World boundaries visualization
function WorldBounds() {
  const geometry = useMemo(() => {
    const geo = new THREE.BoxGeometry(WORLD_SIZE, WORLD_SIZE, 1);
    return geo;
  }, []);

  const edges = useMemo(() => {
    return new THREE.EdgesGeometry(geometry);
  }, [geometry]);

  return (
    <mesh position={[0, 0, -0.5]} rotation={[-Math.PI / 2, 0, 0]}>
      <lineSegments geometry={edges}>
        <lineBasicMaterial color="#666666" linewidth={2} />
      </lineSegments>
    </mesh>
  );
}

// AOI (Area of Interest) visualization
function AOIIndicator({ position, radius, visible }) {
  if (!visible || !position) return null;

  return (
    <mesh position={[position.x, position.y, 0.1]}>
      <ringGeometry args={[radius - 1, radius + 1, 32]} />
      <meshBasicMaterial 
        color="#4a90e2" 
        transparent 
        opacity={0.3}
        side={THREE.DoubleSide}
      />
    </mesh>
  );
}

// Main scene component
export default function WorldScene({ 
  localPlayerPosition, 
  showAOI = false, 
  aoiRadius = 50,
  children 
}) {
  return (
    <div style={{ width: '100vw', height: '100vh' }}>
      <Canvas
        shadows={false} // Disable shadows for performance
        gl={{
          antialias: false, // Disable antialiasing for performance
          powerPreference: 'high-performance'
        }}
        dpr={window.devicePixelRatio * 0.5} // Reduce resolution for performance
      >
        {/* Lighting */}
        <ambientLight intensity={0.8} />
        <directionalLight 
          position={[50, 50, 25]} 
          intensity={0.5}
          castShadow={false}
        />
        
        {/* Camera */}
        <CameraController localPlayerPosition={localPlayerPosition} />
        
        {/* World Grid */}
        <Grid
          args={[WORLD_SIZE, GRID_SIZE]}
          cellSize={WORLD_SIZE / GRID_SIZE}
          cellThickness={0.5}
          cellColor="#2a2a2a"
          sectionSize={WORLD_SIZE / 10}
          sectionThickness={1}
          sectionColor="#444444"
          fadeDistance={WORLD_SIZE / 2}
          fadeStrength={1}
          followCamera={false}
          infiniteGrid={false}
        />
        
        {/* World Boundaries */}
        <WorldBounds />
        
        {/* AOI Indicator */}
        <AOIIndicator 
          position={localPlayerPosition}
          radius={aoiRadius}
          visible={showAOI}
        />
        
        {/* Entity Rendering */}
        {children}
        
        {/* Controls for debugging */}
        <OrbitControls 
          enablePan={true}
          enableZoom={true}
          enableRotate={false} // Disable rotation for top-down view
          minDistance={10}
          maxDistance={200}
          maxPolarAngle={Math.PI / 2.5} // Limit camera angle
        />
      </Canvas>
    </div>
  );
}
