import { useEffect, useRef, useCallback } from 'react';

export const useInput = (onSendInput) => {
  const keysPressed = useRef(new Set());
  const animationFrameId = useRef(null);
  
  // Input state
  const inputState = useRef({
    velocity_x: 0,
    velocity_y: 0
  });

  // Handle keyboard input
  const handleKeyDown = useCallback((event) => {
    // Prevent default for game keys
    if (['w', 'a', 's', 'd', 'arrowup', 'arrowdown', 'arrowleft', 'arrowright'].includes(event.key.toLowerCase())) {
      event.preventDefault();
      keysPressed.current.add(event.key.toLowerCase());
    }
  }, []);

  const handleKeyUp = useCallback((event) => {
    keysPressed.current.delete(event.key.toLowerCase());
  }, []);

  // Process input and send to server
  const processInput = useCallback(() => {
    let velocity_x = 0;
    let velocity_y = 0;
    const speed = 5; // Units per second

    // Calculate velocity based on pressed keys
    if (keysPressed.current.has('w') || keysPressed.current.has('arrowup')) {
      velocity_y += speed;
    }
    if (keysPressed.current.has('s') || keysPressed.current.has('arrowdown')) {
      velocity_y -= speed;
    }
    if (keysPressed.current.has('a') || keysPressed.current.has('arrowleft')) {
      velocity_x -= speed;
    }
    if (keysPressed.current.has('d') || keysPressed.current.has('arrowright')) {
      velocity_x += speed;
    }

    // Normalize diagonal movement
    if (velocity_x !== 0 && velocity_y !== 0) {
      const magnitude = Math.sqrt(velocity_x * velocity_x + velocity_y * velocity_y);
      velocity_x = (velocity_x / magnitude) * speed;
      velocity_y = (velocity_y / magnitude) * speed;
    }

    // Only send input if it changed
    if (velocity_x !== inputState.current.velocity_x || velocity_y !== inputState.current.velocity_y) {
      inputState.current.velocity_x = velocity_x;
      inputState.current.velocity_y = velocity_y;
      
      onSendInput?.(inputState.current);
    }

    // Continue processing
    animationFrameId.current = requestAnimationFrame(processInput);
  }, [onSendInput]);

  // Set up event listeners
  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);

    // Start input processing loop
    animationFrameId.current = requestAnimationFrame(processInput);

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('keyup', handleKeyUp);
      
      if (animationFrameId.current) {
        cancelAnimationFrame(animationFrameId.current);
      }
    };
  }, [handleKeyDown, handleKeyUp, processInput]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (animationFrameId.current) {
        cancelAnimationFrame(animationFrameId.current);
      }
    };
  }, []);

  return {
    getInputState: () => ({ ...inputState.current })
  };
};
