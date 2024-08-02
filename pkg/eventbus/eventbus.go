package eventbus

// Event represents an event with a name and data.
type Event struct {
	Name string
	Data interface{}
}

// EventHandler defines the methods for handling events.
type EventHandler interface {
	HandleEvent(event Event) error
}

// EventBus manages event handlers and dispatching events.
type EventBus struct {
	handlers map[string][]EventHandler
}

// NewEventBus creates a new EventBus instance.
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// RegisterHandler registers an event handler for a specific event.
func (bus *EventBus) RegisterHandler(eventName string, handler EventHandler) {
	bus.handlers[eventName] = append(bus.handlers[eventName], handler)
}

// DispatchEvent dispatches an event to the registered handlers.
func (bus *EventBus) DispatchEvent(event Event) {
	if handlers, found := bus.handlers[event.Name]; found {
		for _, handler := range handlers {
			handler.HandleEvent(event)
		}
	}
}
