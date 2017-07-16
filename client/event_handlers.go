package client

type Event interface{}

type EventHandler func(evt Event) error

type EventHandlers interface {
	Add(EventHandler) error
	Remove(EventHandler) error
	Raise(Event)
}
