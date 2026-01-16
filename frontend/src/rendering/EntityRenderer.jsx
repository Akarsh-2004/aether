import { useRef, useMemo, useEffect } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import useEntityStore from '../state/entityStore';

// Entity geometry cache for performance
const entityGeometry = new THREE.BoxGeometry(2, 2, 1);
const entityMaterial = new THREE.MeshBasicMaterial({ vertexColors: true });

// Individual entity component
function Entity({ entityData }) {
  const meshRef = useRef();
  const { position, velocity, rotation, color, type } = entityData;
  
  // Set initial color
  useEffect(() => {
    if (meshRef.current) {
      meshRef.current.material.color.set(color);
      
      // Different sizes for different entity types
      const scale = type === 'local' ? 1.5 : 1.0;
      meshRef.current.scale.set(scale, scale, scale);
    }
  }, [color, type]);

  // Smooth interpolation between server updates
  useFrame((state, delta) => {
    if (!meshRef.current) return;

    const entity = useEntityStore.getState().getEntity(entityData.id);
    if (!entity) return;

    // Linear interpolation for smooth movement
    const interpolationSpeed = 10; // Adjust for smoothness vs responsiveness
    entity.interpolationTime += delta;

    // Interpolate position
    const targetX = entity.position.x;
    const targetY = entity.position.y;
    const currentX = entity.interpolatedPosition.x;
    const currentY = entity.interpolatedPosition.y;

    const newX = THREE.MathUtils.lerp(currentX, targetX, delta * interpolationSpeed);
    const newY = THREE.MathUtils.lerp(currentY, targetY, delta * interpolationSpeed);

    meshRef.current.position.x = newX;
    meshRef.current.position.y = newY;
    meshRef.current.position.z = 0.5; // Slightly above ground

    // Update interpolated position for next frame
    entity.interpolatedPosition.x = newX;
    entity.interpolatedPosition.y = newY;

    // Smooth rotation
    meshRef.current.rotation.z = THREE.MathUtils.lerp(
      meshRef.current.rotation.z,
      rotation,
      delta * interpolationSpeed
    );

    // Add subtle animation based on velocity
    if (velocity && (velocity.x !== 0 || velocity.y !== 0)) {
      const speed = Math.sqrt(velocity.x * velocity.x + velocity.y * velocity.y);
      const wobble = Math.sin(state.clock.elapsedTime * 10) * 0.05 * Math.min(speed / 10, 1);
      meshRef.current.rotation.z += wobble;
    }
  });

  return (
    <mesh
      ref={meshRef}
      geometry={entityGeometry}
      material={entityMaterial}
      castShadow={false}
      receiveShadow={false}
    />
  );
}

// Instanced entity renderer for performance
function InstancedEntities({ entities }) {
  const instancedMeshRef = useRef();
  
  const { count, positions, colors, scales } = useMemo(() => {
    const count = entities.length;
    const positions = new Float32Array(count * 3);
    const colors = new Float32Array(count * 3);
    const scales = new Float32Array(count * 3);

    entities.forEach((entity, i) => {
      positions[i * 3] = entity.position.x;
      positions[i * 3 + 1] = entity.position.y;
      positions[i * 3 + 2] = 0.5;

      const color = new THREE.Color(entity.color);
      colors[i * 3] = color.r;
      colors[i * 3 + 1] = color.g;
      colors[i * 3 + 2] = color.b;

      const scale = entity.type === 'local' ? 1.5 : 1.0;
      scales[i * 3] = scale;
      scales[i * 3 + 1] = scale;
      scales[i * 3 + 2] = scale;
    });

    return { count, positions, colors, scales };
  }, [entities]);

  useFrame((state, delta) => {
    if (!instancedMeshRef.current) return;

    const tempPosition = new THREE.Object3D();
    const interpolationSpeed = 10;

    entities.forEach((entity, i) => {
      // Interpolate position
      const targetX = entity.position.x;
      const targetY = entity.position.y;
      const currentX = entity.interpolatedPosition.x;
      const currentY = entity.interpolatedPosition.y;

      const newX = THREE.MathUtils.lerp(currentX, targetX, delta * interpolationSpeed);
      const newY = THREE.MathUtils.lerp(currentY, targetY, delta * interpolationSpeed);

      entity.interpolatedPosition.x = newX;
      entity.interpolatedPosition.y = newY;

      tempPosition.position.set(newX, newY, 0.5);
      tempPosition.rotation.z = entity.rotation;
      
      const scale = entity.type === 'local' ? 1.5 : 1.0;
      tempPosition.scale.set(scale, scale, scale);
      
      tempPosition.updateMatrix();
      instancedMeshRef.current.setMatrixAt(i, tempPosition.matrix);
    });

    instancedMeshRef.current.instanceMatrix.needsUpdate = true;
  });

  return (
    <instancedMesh
      ref={instancedMeshRef}
      args={[entityGeometry, entityMaterial, count]}
      castShadow={false}
      receiveShadow={false}
    >
      <instancedBufferAttribute
        attach="attributes-color"
        count={count}
        array={colors}
        itemSize={3}
      />
    </instancedMesh>
  );
}

// Main entity renderer component
export default function EntityRenderer() {
  const entities = useEntityStore(state => state.getVisibleEntities());
  const localPlayerId = useEntityStore(state => state.localPlayerId);
  
  // Separate local player from other entities for special handling
  const localPlayer = entities.find(e => e.id === localPlayerId);
  const otherEntities = entities.filter(e => e.id !== localPlayerId);

  // Use instanced rendering for better performance with many entities
  const useInstancedRendering = otherEntities.length > 50;

  return (
    <>
      {/* Render local player separately */}
      {localPlayer && <Entity key={localPlayer.id} entityData={localPlayer} />}
      
      {/* Render other entities */}
      {useInstancedRendering ? (
        <InstancedEntities entities={otherEntities} />
      ) : (
        otherEntities.map(entity => (
          <Entity key={entity.id} entityData={entity} />
        ))
      )}
    </>
  );
}
