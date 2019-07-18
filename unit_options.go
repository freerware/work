package work

import (
	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

// UnitOptions represents the configuration options
// for the work unit.
type UnitOptions struct {
	Logger *zap.Logger
	Scope  tally.Scope
}

// Option applies an option to the provided configuration.
type Option func(*UnitOptions)

var (
	// Logger specifies the option to provide a logger for the work unit.
	UnitLogger = func(l *zap.Logger) Option {
		return func(o *UnitOptions) {
			o.Logger = l
		}
	}

	// Scope specifies the option to provide a metric scope for the work unit.
	UnitScope = func(s tally.Scope) Option {
		return func(o *UnitOptions) {
			o.Scope = s
		}
	}
)
