package profiletometrics

import (
	"context"
	"testing"

	"github.com/henrikrexed/profiletoMetrics/pkg/profiletometrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer/consumertest"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	assert.NotNil(t, factory)
	assert.Equal(t, component.MustNewType("profiletometrics"), factory.Type())
}

func TestCreateDefaultConfig(t *testing.T) {
	config := createDefaultConfig()
	assert.NotNil(t, config)

	cfg, ok := config.(*Config)
	require.True(t, ok)
	assert.True(t, cfg.ConverterConfig.Metrics.CPU.Enabled)
	assert.True(t, cfg.ConverterConfig.Metrics.Memory.Enabled)
	assert.Equal(t, "cpu_time", cfg.ConverterConfig.Metrics.CPU.Name)
	assert.Equal(t, "memory_allocation", cfg.ConverterConfig.Metrics.Memory.Name)
}

func TestCreateProfilesToMetricsConnector(t *testing.T) {
	settings := connector.Settings{
		ID:                component.NewID(component.MustNewType("profiletometrics")),
		TelemetrySettings: componenttest.NewNopTelemetrySettings(),
		BuildInfo:         component.NewDefaultBuildInfo(),
	}

	config := &Config{
		ConverterConfig: profiletometrics.ConverterConfig{
			Metrics: profiletometrics.MetricsConfig{
				CPU: profiletometrics.CPUMetricConfig{
					Enabled: true,
					Name:    "cpu_time",
					Unit:    "ns",
				},
			},
		},
	}

	nextConsumer := consumertest.NewNop()

	connector, err := createProfilesToMetricsConnector(
		context.Background(),
		settings,
		config,
		nextConsumer,
	)

	assert.NoError(t, err)
	assert.NotNil(t, connector)
}
