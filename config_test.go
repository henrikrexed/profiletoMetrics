package profiletometrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/henrikrexed/profiletoMetrics/pkg/profiletometrics"
)

func TestConfig_Structure(t *testing.T) {
	config := &Config{
		ConverterConfig: profiletometrics.ConverterConfig{
			Metrics: profiletometrics.MetricsConfig{
				CPU: profiletometrics.CPUMetricConfig{
					Enabled: true,
					Name:    "cpu_time",
					Unit:    "ns",
				},
				Memory: profiletometrics.MemoryMetricConfig{
					Enabled: true,
					Name:    "memory_allocation",
					Unit:    "bytes",
				},
			},
		},
	}

	assert.True(t, config.ConverterConfig.Metrics.CPU.Enabled)
	assert.True(t, config.ConverterConfig.Metrics.Memory.Enabled)
	assert.Equal(t, "cpu_time", config.ConverterConfig.Metrics.CPU.Name)
	assert.Equal(t, "memory_allocation", config.ConverterConfig.Metrics.Memory.Name)
}

func TestConfig_Empty(t *testing.T) {
	config := &Config{}
	assert.NotNil(t, config)
}
