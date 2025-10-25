package profiletometrics

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/connector/xconnector"
	"go.opentelemetry.io/collector/consumer"

	"github.com/henrikrexed/profiletoMetrics/pkg/profiletometrics"
)

const (
	typeStr = "profiletometrics"
)

// NewFactory creates a new connector factory
func NewFactory() connector.Factory {
	return xconnector.NewFactory(
		component.MustNewType(typeStr),
		createDefaultConfig,
		// This is a pure profiles-to-metrics connector
		xconnector.WithProfilesToMetrics(createProfilesToMetricsConnector, component.StabilityLevelAlpha),
	)
}

func createProfilesToMetricsConnector(
	_ context.Context,
	set connector.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (xconnector.Profiles, error) {
	config := cfg.(*Config)
	converter, err := profiletometrics.NewConverter(&config.ConverterConfig)
	if err != nil {
		return nil, err
	}

	// Set the logger on the converter
	converter.SetLogger(set.Logger)

	return &profileToMetricsConnector{
		config:       config,
		nextConsumer: nextConsumer,
		logger:       set.Logger,
		converter:    converter,
	}, nil
}

func createDefaultConfig() component.Config {
	return &Config{
		ConverterConfig: profiletometrics.ConverterConfig{
			Metrics: profiletometrics.MetricsConfig{
				CPU: profiletometrics.CPUMetricConfig{
					Enabled:    true,
					MetricName: "cpu_time",
					Unit:       "ns",
				},
				Memory: profiletometrics.MemoryMetricConfig{
					Enabled:    true,
					MetricName: "memory_allocation",
					Unit:       "bytes",
				},
				Function: profiletometrics.FunctionMetricConfig{
					Enabled: true,
				},
			},
			Attributes: []profiletometrics.AttributeConfig{
				{
					Key:   "service.name",
					Value: "service_name",
					Type:  "literal",
				},
				{
					Key:   "process.name",
					Value: "process_name",
					Type:  "literal",
				},
				{
					Key:   "function.name",
					Value: "function_name",
					Type:  "regex",
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
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate that at least one metric is enabled
	if !c.ConverterConfig.Metrics.CPU.Enabled && !c.ConverterConfig.Metrics.Memory.Enabled {
		return fmt.Errorf("at least one metric must be enabled")
	}
	return nil
}
