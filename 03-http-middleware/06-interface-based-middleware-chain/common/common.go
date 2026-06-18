package common

import (
	"context"
)

type Event struct {
	ID   string
	Type string
	Message string 
}


type Processor interface {
	Process(context.Context, Event) ([]Event, error)
}