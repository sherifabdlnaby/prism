package component

import (
	"github.com/sherifabdlnaby/prism/pkg/config"
	"go.uber.org/zap"
)

// Base defines the basic prism component.
type Base interface {
	// Init Initializes Base's configuration
	Init(config.Config, zap.SugaredLogger) error

	// start starts the component
	Start() error

	// Stop shutdown down and clean up resources gracefully within a timeout.
	Close() error
}
