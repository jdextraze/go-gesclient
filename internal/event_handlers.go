package internal

import (
	"github.com/jdextraze/go-gesclient/models"
	"fmt"
	"errors"
	"reflect"
)

type eventHandlers struct {
	handlers []models.EventHandler
}

func newEventHandlers() *eventHandlers {
	return &eventHandlers{
		handlers: make([]models.EventHandler, 0, 10),
	}
}

func (h *eventHandlers) Add(handler models.EventHandler) error {
	h.handlers = append(h.handlers, handler)
	return nil
}

func (h *eventHandlers) Remove(handler models.EventHandler) error {
	pos := -1
	for i, h := range h.handlers {
		if fmt.Sprintf("%v", h) == fmt.Sprintf("%v", handler) {
			pos = i
			break
		}
	}
	if pos == -1 {
		return errors.New("handler not found")
	}
	h.handlers = append(h.handlers[:pos], h.handlers[pos+1:]...)
	return nil
}

func (h *eventHandlers) Raise(evt models.Event) {
	go func() {
		for _, h := range h.handlers {
			if err := h(evt); err != nil {
				log.Errorf("Error occurred while raising event %s: %v", reflect.TypeOf(evt), err)
			}
		}
	}()
}
