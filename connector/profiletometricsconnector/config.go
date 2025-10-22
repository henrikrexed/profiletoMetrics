package profiletometricsconnector

import (
	"github.com/henrikrexed/profiletoMetrics/pkg/profiletometrics"
)

// Config defines the configuration for the ProfileToMetrics connector.
type Config struct {
	// Metrics configuration
	Metrics profiletometrics.MetricsConfig `mapstructure:"metrics"`

	// Attributes to extract from profiles
	Attributes []profiletometrics.AttributeConfig `mapstructure:"attributes"`

	// Process filtering configuration
	ProcessFilter profiletometrics.ProcessFilterConfig `mapstructure:"process_filter"`

	// Pattern filtering configuration
	PatternFilter profiletometrics.PatternFilterConfig `mapstructure:"pattern_filter"`

	// Thread filtering configuration
	ThreadFilter profiletometrics.ThreadFilterConfig `mapstructure:"thread_filter"`
}
