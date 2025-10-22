package profiletometrics

import (
	"github.com/henrikrexed/profiletoMetrics/connector"
	otelconnector "go.opentelemetry.io/collector/connector"
)

// NewFactory creates a new connector factory
func NewFactory() otelconnector.Factory {
	return connector.NewFactory()
}
