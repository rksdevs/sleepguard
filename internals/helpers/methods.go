package helpers

import "fmt"

func (e *Events) ParsedEvent() {
	fmt.Printf("This is an event of %s from %s with a state of %s", e.Type, e.Source, e.State)
}