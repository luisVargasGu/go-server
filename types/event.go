package types

import (
	"log"
	"sync"
)

const (
	EventRegister        = "register"
	EventUnregister      = "unregister"
	EventBroadcast       = "broadcast"
	EventWebRTCSignaling = "webrtc-signaling"
	EventChatMessage     = "chat-message"
)

type Event struct {
	Type    string
	Payload interface{}
}

type EventBus struct {
	subscribers map[string][]chan Event
	mu          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan Event),
	}
}

func (bus *EventBus) Subscribe(eventType string, ch chan Event) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.subscribers[eventType] = append(bus.subscribers[eventType], ch)
}

func (bus *EventBus) Publish(event Event) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	for _, ch := range bus.subscribers[event.Type] {
		select {
		case ch <- event:
		default:
			log.Printf("Subscriber channel full for event type: %s", event.Type)
		}
	}
}
