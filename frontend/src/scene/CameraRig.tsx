import React, { useRef } from 'react';
import { useFrame, useThree } from '@react-three/fiber';
import { OrbitControls } from '@react-three/drei';
import * as THREE from 'three';

interface CameraRigProps {
  targetPosition?: [number, number, number];
  enableOrbit?: boolean;
  maxDistance?: number;
  minDistance?: number;
}

/**
 * Camera rig for isometric/simulation-style viewing
 * 
 * Design rationale:
 * - Elevated isometric angle for tactical overview
 * - Restricted rotation prevents disorientation
 * - Smooth panning and zoom for navigation
 * - Focused on simulation use, not gaming
 */
export const CameraRig = React.memo<CameraRigProps>(({
  targetPosition = [0, 0, 0],
  enableOrbit = true,
  maxDistance = 80,
  minDistance = 20
}) => {
  const controlsRef = useRef<any>(null);
  const { camera, gl } = useThree();

  // Set up camera when component mounts
  React.useEffect(() => {
    if (camera instanceof THREE.PerspectiveCamera) {
      camera.position.set(30, 25, 30);
      camera.fov = 45;
      camera.near = 0.1;
      camera.far = 1000;
      camera.updateProjectionMatrix();
    }
  }, [camera]);

  // Smooth camera following for dynamic target tracking
  useFrame((state) => {
    if (controlsRef.current && targetPosition) {
      // Smoothly interpolate camera target
      const currentTarget = controlsRef.current.target;
      currentTarget.x += (targetPosition[0] - currentTarget.x) * 0.05;
      currentTarget.y += (targetPosition[1] - currentTarget.y) * 0.05;
      currentTarget.z += (targetPosition[2] - currentTarget.z) * 0.05;
    }
  });

  return (
    <OrbitControls
      ref={controlsRef}
      args={[camera, gl.domElement]}
      enablePan={true}
      enableZoom={true}
      enableRotate={enableOrbit}
      maxPolarAngle={Math.PI * 0.6} // Limit vertical rotation
      minPolarAngle={Math.PI * 0.2}
      maxDistance={maxDistance}
      minDistance={minDistance}
      enableDamping={true}
      dampingFactor={0.05}
      screenSpacePanning={false} // Pan in world space, not screen space
      target={new THREE.Vector3(...targetPosition)}
    />
  );
});

CameraRig.displayName = 'CameraRig';
