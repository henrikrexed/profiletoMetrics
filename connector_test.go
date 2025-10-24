package profiletometrics

import (
	"context"
	"testing"

	"github.com/henrikrexed/profiletoMetrics/pkg/profiletometrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pprofile"
)

func TestProfileToMetricsConnector_Start(t *testing.T) {
	connector := &profileToMetricsConnector{
		config:       &Config{},
		nextConsumer: consumertest.NewNop(),
		logger:       componenttest.NewNopTelemetrySettings().Logger,
		converter:    nil, // Will be set in Start
	}

	host := componenttest.NewNopHost()
	err := connector.Start(context.Background(), host)
	assert.NoError(t, err)
}

func TestProfileToMetricsConnector_Shutdown(t *testing.T) {
	connector := &profileToMetricsConnector{
		config:       &Config{},
		nextConsumer: consumertest.NewNop(),
		logger:       componenttest.NewNopTelemetrySettings().Logger,
		converter:    nil,
	}

	err := connector.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestProfileToMetricsConnector_Capabilities(t *testing.T) {
	connector := &profileToMetricsConnector{
		config:       &Config{},
		nextConsumer: consumertest.NewNop(),
		logger:       componenttest.NewNopTelemetrySettings().Logger,
		converter:    nil,
	}

	capabilities := connector.Capabilities()
	assert.True(t, capabilities.MutatesData)
}

func TestProfileToMetricsConnector_ConsumeProfiles(t *testing.T) {
	// Create a mock converter
	converter, err := profiletometrics.NewConverter(&profiletometrics.ConverterConfig{
		Metrics: profiletometrics.MetricsConfig{
			CPU: profiletometrics.CPUMetricConfig{
				Enabled:    true,
				MetricName: "cpu_time",
				Unit:       "ns",
			},
		},
	})
	require.NoError(t, err)

	connector := &profileToMetricsConnector{
		config:       &Config{},
		nextConsumer: consumertest.NewNop(),
		logger:       componenttest.NewNopTelemetrySettings().Logger,
		converter:    converter,
	}

	// Create empty profiles
	profiles := pprofile.NewProfiles()

	err = connector.ConsumeProfiles(context.Background(), profiles)
	assert.NoError(t, err)
}
