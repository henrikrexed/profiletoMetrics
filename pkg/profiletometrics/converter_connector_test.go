package profiletometrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/pprofile"
)

func TestNewConverterConnector(t *testing.T) {
	config := ConverterConfig{
		Metrics: MetricsConfig{
			CPU: CPUMetricConfig{
				Enabled: true,
				Name:    "test_cpu",
				Unit:    "s",
			},
		},
	}

	connector := NewConverterConnector(config)
	assert.NotNil(t, connector)
	assert.Equal(t, config, connector.config)
	assert.NotNil(t, connector.logger)
}

func TestConverterConnector_ConvertTracesToMetrics(t *testing.T) {
	config := ConverterConfig{
		Metrics: MetricsConfig{
			CPU: CPUMetricConfig{
				Enabled: true,
				Name:    "test_cpu",
				Unit:    "s",
			},
		},
	}

	connector := NewConverterConnector(config)
	traces := ptrace.NewTraces()

	metrics, err := connector.ConvertTracesToMetrics(traces)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.ResourceMetrics().Len())
}

func TestConverterConnector_ConvertLogsToMetrics(t *testing.T) {
	config := ConverterConfig{
		Metrics: MetricsConfig{
			CPU: CPUMetricConfig{
				Enabled: true,
				Name:    "test_cpu",
				Unit:    "s",
			},
		},
	}

	connector := NewConverterConnector(config)
	logs := plog.NewLogs()

	metrics, err := connector.ConvertLogsToMetrics(logs)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.ResourceMetrics().Len())
}

func TestConverterConnector_ConvertProfilesToMetrics(t *testing.T) {
	config := ConverterConfig{
		Metrics: MetricsConfig{
			CPU: CPUMetricConfig{
				Enabled: true,
				Name:    "test_cpu",
				Unit:    "s",
			},
		},
	}

	connector := NewConverterConnector(config)
	profiles := pprofile.NewProfiles()

	metrics, err := connector.ConvertProfilesToMetrics(context.Background(), profiles)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, metrics.ResourceMetrics().Len())
}
