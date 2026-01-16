package protocol

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
	"github.com/akarsh-2004/aether/proto"
)

var (
	ErrInvalidMessage = errors.New("invalid message")
	ErrUnknownType    = errors.New("unknown message type")
)

type Codec struct{}

func NewCodec() *Codec {
	return &Codec{}
}

// Encode marshals a protobuf message to binary format
func (c *Codec) Encode(msg *proto.Message) ([]byte, error) {
	if msg == nil {
		return nil, ErrInvalidMessage
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

// Decode unmarshals binary data to a protobuf message
func (c *Codec) Decode(data []byte) (*proto.Message, error) {
	if len(data) == 0 {
		return nil, ErrInvalidMessage
	}

	var msg proto.Message
	if err := proto.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &msg, nil
}

// ValidateMessage performs basic validation on a message
func (c *Codec) ValidateMessage(msg *proto.Message) error {
	if msg == nil {
		return ErrInvalidMessage
	}

	switch msg.Type {
	case proto.MessageType_MOVEMENT_DELTA:
		if msg.MovementDelta == nil {
			return fmt.Errorf("movement_delta payload is required for MOVEMENT_DELTA type")
		}
		if msg.MovementDelta.EntityId == 0 {
			return fmt.Errorf("entity_id is required in movement_delta")
		}

	case proto.MessageType_ENTITY_STATE:
		if msg.EntityState == nil {
			return fmt.Errorf("entity_state payload is required for ENTITY_STATE type")
		}
		if msg.EntityState.EntityId == 0 {
			return fmt.Errorf("entity_id is required in entity_state")
		}

	case proto.MessageType_SERVER_SNAPSHOT:
		if msg.ServerSnapshot == nil {
			return fmt.Errorf("server_snapshot payload is required for SERVER_SNAPSHOT type")
		}

	case proto.MessageType_SPAWN_REQUEST:
		if msg.SpawnRequest == nil {
			return fmt.Errorf("spawn_request payload is required for SPAWN_REQUEST type")
		}
		if msg.SpawnRequest.ClientId == "" {
			return fmt.Errorf("client_id is required in spawn_request")
		}

	case proto.MessageType_SPAWN_RESPONSE:
		if msg.SpawnResponse == nil {
			return fmt.Errorf("spawn_response payload is required for SPAWN_RESPONSE type")
		}

	case proto.MessageType_CORRECTION:
		if msg.Correction == nil {
			return fmt.Errorf("correction payload is required for CORRECTION type")
		}
		if msg.Correction.EntityId == 0 {
			return fmt.Errorf("entity_id is required in correction")
		}

	case proto.MessageType_DESPAWN:
		if msg.Despawn == nil {
			return fmt.Errorf("despawn payload is required for DESPAWN type")
		}
		if msg.Despawn.EntityId == 0 {
			return fmt.Errorf("entity_id is required in despawn")
		}

	case proto.MessageType_HEARTBEAT:
		if msg.Heartbeat == nil {
			return fmt.Errorf("heartbeat payload is required for HEARTBEAT type")
		}
		if msg.Heartbeat.ClientId == "" {
			return fmt.Errorf("client_id is required in heartbeat")
		}

	default:
		return ErrUnknownType
	}

	return nil
}
