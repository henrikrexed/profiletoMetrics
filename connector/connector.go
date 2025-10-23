package connector

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.uber.org/zap"

	"github.com/henrikrexed/profiletoMetrics/pkg/profiletometrics"
)

// profileToMetricsConnector implements the ProfileToMetrics connector.
type profileToMetricsConnector struct {
	config       *Config
	nextConsumer consumer.Metrics
	logger       *zap.Logger
	converter    *profiletometrics.Converter
}

// Start implements component.Component.
func (c *profileToMetricsConnector) Start(ctx context.Context, host component.Host) error {
	c.logger.Info("Starting ProfileToMetrics connector")

	// Log configuration details at debug level
	c.logger.Debug("ProfileToMetrics connector configuration",
		zap.Any("metrics_config", c.config.Metrics),
		zap.Any("attributes", c.config.Attributes),
		zap.Any("process_filter", c.config.ProcessFilter),
		zap.Any("pattern_filter", c.config.PatternFilter),
		zap.Any("thread_filter", c.config.ThreadFilter),
	)

	// Initialize the converter with the configuration
	converterConfig := &profiletometrics.ConverterConfig{
		Metrics:       c.config.Metrics,
		Attributes:    c.config.Attributes,
		ProcessFilter: c.config.ProcessFilter,
		PatternFilter: c.config.PatternFilter,
		ThreadFilter:  c.config.ThreadFilter,
	}

	var err error
	c.converter, err = profiletometrics.NewConverter(converterConfig)
	if err != nil {
		c.logger.Error("Failed to initialize converter", zap.Error(err))
		return err
	}

	c.logger.Debug("ProfileToMetrics connector initialized successfully")
	return nil
}

// Shutdown implements component.Component.
func (c *profileToMetricsConnector) Shutdown(ctx context.Context) error {
	c.logger.Info("Shutting down ProfileToMetrics connector")
	c.logger.Debug("ProfileToMetrics connector shutdown completed")
	return nil
}

// Capabilities implements connector interfaces.
func (c *profileToMetricsConnector) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

// ConsumeProfiles implements connector.Profiles.
func (c *profileToMetricsConnector) ConsumeProfiles(ctx context.Context, profiles pprofile.Profiles) error {
	// Log input statistics
	resourceProfilesCount := profiles.ResourceProfiles().Len()
	totalSamples := profiles.SampleCount()

	c.logger.Debug("Processing profiles",
		zap.Int("resource_profiles_count", resourceProfilesCount),
		zap.Int("total_samples", totalSamples),
	)

	// Convert profiles to metrics using the converter
	metrics, err := c.converter.ConvertProfilesToMetrics(ctx, profiles)
	if err != nil {
		c.logger.Error("Failed to convert profiles to metrics",
			zap.Error(err),
			zap.Int("input_samples", totalSamples),
		)
		return err
	}

	// Log output statistics
	resourceMetricsCount := metrics.ResourceMetrics().Len()
	totalMetrics := 0
	for i := 0; i < resourceMetricsCount; i++ {
		scopeMetrics := metrics.ResourceMetrics().At(i).ScopeMetrics()
		for j := 0; j < scopeMetrics.Len(); j++ {
			totalMetrics += scopeMetrics.At(j).Metrics().Len()
		}
	}

	c.logger.Debug("Profiles converted to metrics",
		zap.Int("input_samples", totalSamples),
		zap.Int("output_resource_metrics", resourceMetricsCount),
		zap.Int("output_metrics", totalMetrics),
	)

	// Send metrics to the next consumer
	if err := c.nextConsumer.ConsumeMetrics(ctx, metrics); err != nil {
		c.logger.Error("Failed to send metrics to next consumer",
			zap.Error(err),
			zap.Int("metrics_count", totalMetrics),
		)
		return err
	}

	c.logger.Debug("Profiles successfully processed and metrics sent to next consumer")
	return nil
}
