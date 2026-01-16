import * as protobuf from 'protobufjs';

// Load protobuf definition asynchronously
let root = null;
let WorldSnapshot = null;
let EntityState = null;
let MovementDelta = null;
let ClientInput = null;

export async function initProtobuf() {
  try {
    // Create protobuf definition programmatically
    root = new protobuf.Root();
    
    // Define Vec2 message
    const Vec2 = new protobuf.Type("Vec2")
      .add(new protobuf.Field("x", 1, "float"))
      .add(new protobuf.Field("y", 2, "float"));
    
    // Define EntityState message
    const EntityState = new protobuf.Type("EntityState")
      .add(new protobuf.Field("id", 1, "string"))
      .add(new protobuf.Field("position", 2, "Vec2"))
      .add(new protobuf.Field("velocity", 3, "Vec2"))
      .add(new protobuf.Field("rotation", 4, "float"))
      .add(new protobuf.Field("lastUpdate", 5, "int64"));
    
    // Define MovementDelta message
    const MovementDelta = new protobuf.Type("MovementDelta")
      .add(new protobuf.Field("id", 1, "string"))
      .add(new protobuf.Field("position", 2, "Vec2"))
      .add(new protobuf.Field("velocity", 3, "Vec2"))
      .add(new protobuf.Field("rotation", 4, "float"))
      .add(new protobuf.Field("timestamp", 5, "int64"));
    
    // Define ClientInput message
    const ClientInput = new protobuf.Type("ClientInput")
      .add(new protobuf.Field("velocity_x", 1, "float"))
      .add(new protobuf.Field("velocity_y", 2, "float"));
    
    // Define WorldSnapshot message
    const WorldSnapshot = new protobuf.Type("WorldSnapshot")
      .add(new protobuf.Field("entities", 1, "EntityState", "repeated"))
      .add(new protobuf.Field("deltas", 2, "MovementDelta", "repeated"));

    // Add types to root
    root.add(Vec2).add(EntityState).add(MovementDelta).add(ClientInput).add(WorldSnapshot);
    
    // Resolve all types
    root.resolveAll();

    return true;
  } catch (error) {
    console.error('Failed to initialize protobuf:', error);
    return false;
  }
}

export function decodeWorldSnapshot(buffer) {
  if (!root) {
    throw new Error('Protobuf not initialized');
  }
  
  try {
    const WorldSnapshot = root.lookupType('WorldSnapshot');
    const message = WorldSnapshot.decode(buffer);
    return WorldSnapshot.toObject(message);
  } catch (error) {
    console.error('Failed to decode WorldSnapshot:', error);
    throw error;
  }
}

export function encodeClientInput(input) {
  if (!root) {
    throw new Error('Protobuf not initialized');
  }
  
  try {
    const ClientInput = root.lookupType('ClientInput');
    const message = ClientInput.create(input);
    return ClientInput.encode(message).finish();
  } catch (error) {
    console.error('Failed to encode ClientInput:', error);
    throw error;
  }
}
