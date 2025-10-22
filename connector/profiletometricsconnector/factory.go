package profiletometricsconnector

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"

	"github.com/henrikrexed/profiletoMetrics/pkg/profiletometrics"
)

const (
	typeStr   = "profiletometrics"
	stability = component.StabilityLevelAlpha
)

var (
	// typeStrComponent is the component type for the profiletometrics connector
	typeStrComponent = component.MustNewType(typeStr)
)

// NewFactory creates a new ProfileToMetrics connector factory.
func NewFactory() connector.Factory {
	return connector.NewFactory(
		typeStrComponent,
		createDefaultConfig,
		connector.WithTracesToMetrics(createTracesToMetricsConnector, stability),
		connector.WithLogsToMetrics(createLogsToMetricsConnector, stability),
		connector.WithMetricsToMetrics(createMetricsToMetricsConnector, stability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
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
		Attributes: []profiletometrics.AttributeConfig{
			{
				Name:  "service.name",
				Value: "service_name",
				Type:  "literal",
			},
		},
		ProcessFilter: profiletometrics.ProcessFilterConfig{
			Enabled: false,
		},
		PatternFilter: profiletometrics.PatternFilterConfig{
			Enabled: false,
		},
		ThreadFilter: profiletometrics.ThreadFilterConfig{
			Enabled: false,
		},
	}
}

func createTracesToMetricsConnector(
	_ context.Context,
	set connector.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (connector.Traces, error) {
	config := cfg.(*Config)
	return &profileToMetricsConnector{
		config:       config,
		nextConsumer: nextConsumer,
		logger:       set.Logger,
	}, nil
}

func createLogsToMetricsConnector(
	_ context.Context,
	set connector.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (connector.Logs, error) {
	config := cfg.(*Config)
	return &profileToMetricsConnector{
		config:       config,
		nextConsumer: nextConsumer,
		logger:       set.Logger,
	}, nil
}

func createMetricsToMetricsConnector(
	_ context.Context,
	set connector.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (connector.Metrics, error) {
	config := cfg.(*Config)
	return &profileToMetricsConnector{
		config:       config,
		nextConsumer: nextConsumer,
		logger:       set.Logger,
	}, nil
}
