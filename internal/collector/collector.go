package collector

import (
	"context"

	"github.com/ITlusions/ITL.Amalia.EdgeSensor/internal/event"
)

type Collector interface {
	Name() string
	Collect(context.Context) ([]event.Event, error)
}
