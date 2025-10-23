package profiletometrics

import (
	"github.com/henrikrexed/profiletoMetrics/pkg/profiletometrics"
)

// Config defines the configuration for the profiletometrics connector
type Config struct {
	// ConverterConfig embeds the converter configuration
	ConverterConfig profiletometrics.ConverterConfig `mapstructure:",squash"`
}
