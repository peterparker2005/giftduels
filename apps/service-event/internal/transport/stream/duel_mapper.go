package stream

import (
	"fmt"
	"strings"

	"github.com/ThreeDotsLabs/watermill/message"
	duelv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/duel/v1"
	eventv1 "github.com/peterparker2005/giftduels/packages/protobuf-go/gen/giftduels/event/v1"
	"google.golang.org/protobuf/proto"
)

func DuelMessageMapper(topic string, msg *message.Message) (*eventv1.StreamResponse, error) {
	// Handle different topic formats
	if topic == "duel.created" {
		var ev duelv1.DuelCreatedEvent
		if err := proto.Unmarshal(msg.Payload, &ev); err != nil {
			return nil, fmt.Errorf("unmarshal DuelCreatedEvent: %w", err)
		}
		return &eventv1.StreamResponse{
			Payload: &eventv1.StreamResponse_DuelEvent{
				DuelEvent: &duelv1.DuelEvent{
					Event: &duelv1.DuelEvent_DuelCreatedEvent{
						DuelCreatedEvent: &ev,
					},
				},
			},
		}, nil
	}

	// Handle "duel:<id>" format
	parts := strings.SplitN(topic, ":", 2)
	if len(parts) != 2 || parts[0] != "duel" {
		return nil, ErrNotOurTopic
	}

	var ev duelv1.DuelEvent
	if err := proto.Unmarshal(msg.Payload, &ev); err != nil {
		return nil, fmt.Errorf("unmarshal DuelEvent: %w", err)
	}
	return &eventv1.StreamResponse{
		Payload: &eventv1.StreamResponse_DuelEvent{
			DuelEvent: &ev,
		},
	}, nil
}
