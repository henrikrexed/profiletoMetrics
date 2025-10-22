package connector

import (
	"context"

	"github.com/henrikrexed/profiletoMetrics/pkg/profiletometrics"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.uber.org/zap"
)

const (
	typeStr = "profiletometrics"
)

// NewFactory creates a new connector factory
func NewFactory() connector.Factory {
	return connector.NewFactory(
		component.MustNewType(typeStr),
		createDefaultConfig,
		// Note: We'll need to implement a custom profiles-to-metrics connector
		// since the standard API doesn't have this yet
	)
}

func createDefaultConfig() component.Config {
	return &profiletometrics.ConverterConfig{
		Metrics: profiletometrics.MetricsConfig{
			CPU: profiletometrics.CPUMetricConfig{
				Enabled: true,
				Name:    "cpu_time",
				Unit:    "s",
			},
			Memory: profiletometrics.MemoryMetricConfig{
				Enabled: true,
				Name:    "memory_allocation",
				Unit:    "bytes",
			},
		},
	}
}

// ProfilesToMetricsConnector implements a custom profiles-to-metrics connector
type ProfilesToMetricsConnector struct {
	config       profiletometrics.ConverterConfig
	logger       *zap.Logger
	nextConsumer consumer.Metrics
}

// NewProfilesToMetricsConnector creates a new profiles-to-metrics connector
func NewProfilesToMetricsConnector(config profiletometrics.ConverterConfig, nextConsumer consumer.Metrics) *ProfilesToMetricsConnector {
	logger, _ := zap.NewDevelopment()
	return &ProfilesToMetricsConnector{
		config:       config,
		logger:       logger,
		nextConsumer: nextConsumer,
	}
}

// ConsumeProfiles processes profiles and converts them to metrics
func (c *ProfilesToMetricsConnector) ConsumeProfiles(ctx context.Context, profiles pprofile.Profiles) error {
	c.logger.Debug("Processing profiles",
		zap.Int("resource_profiles_count", profiles.ResourceProfiles().Len()),
		zap.Int("sample_count", profiles.SampleCount()),
	)

	// Convert profiles to metrics using our converter
	converter := profiletometrics.NewConverterConnector(c.config)
	metrics, err := converter.ConvertProfilesToMetrics(ctx, profiles)
	if err != nil {
		c.logger.Error("Failed to convert profiles to metrics", zap.Error(err))
		return err
	}

	// Send metrics to the next consumer
	if err := c.nextConsumer.ConsumeMetrics(ctx, metrics); err != nil {
		c.logger.Error("Failed to consume metrics", zap.Error(err))
		return err
	}

	c.logger.Debug("Successfully processed profiles to metrics",
		zap.Int("output_metrics", metrics.ResourceMetrics().Len()),
	)

	return nil
}

// Capabilities returns the capabilities of this connector
func (c *ProfilesToMetricsConnector) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

// Start initializes the connector
func (c *ProfilesToMetricsConnector) Start(_ context.Context, _ component.Host) error {
	c.logger.Info("Starting profiles-to-metrics connector")
	return nil
}

// Shutdown stops the connector
func (c *ProfilesToMetricsConnector) Shutdown(_ context.Context) error {
	c.logger.Info("Shutting down profiles-to-metrics connector")
	return nil
}
