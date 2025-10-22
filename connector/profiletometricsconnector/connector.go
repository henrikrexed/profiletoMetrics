package profiletometricsconnector

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"

	"github.com/henrikrexed/profiletoMetrics/pkg/profiletometrics"
)

// profileToMetricsConnector implements the ProfileToMetrics connector.
type profileToMetricsConnector struct {
	config       *Config
	nextConsumer consumer.Metrics
	logger       *zap.Logger
	converter    *profiletometrics.ConverterConnector
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
	converterConfig := profiletometrics.ConverterConfig{
		Metrics:       c.config.Metrics,
		Attributes:    c.config.Attributes,
		ProcessFilter: c.config.ProcessFilter,
		PatternFilter: c.config.PatternFilter,
		ThreadFilter:  c.config.ThreadFilter,
	}
	c.converter = profiletometrics.NewConverterConnector(converterConfig)

	c.logger.Debug("ProfileToMetrics connector initialized successfully")
	return nil
}

// Shutdown implements component.Component.
func (c *profileToMetricsConnector) Shutdown(ctx context.Context) error {
	c.logger.Info("Shutting down ProfileToMetrics connector")
	c.logger.Debug("ProfileToMetrics connector shutdown completed")
	return nil
}

// Capabilities implements connector.Traces.
func (c *profileToMetricsConnector) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

// ConsumeTraces implements connector.Traces.
func (c *profileToMetricsConnector) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	// Log input statistics
	resourceSpansCount := td.ResourceSpans().Len()
	totalSpans := 0
	for i := 0; i < resourceSpansCount; i++ {
		totalSpans += td.ResourceSpans().At(i).ScopeSpans().Len()
	}

	c.logger.Debug("Processing traces",
		zap.Int("resource_spans_count", resourceSpansCount),
		zap.Int("total_spans", totalSpans),
	)

	// Convert traces to metrics using the converter
	metrics, err := c.converter.ConvertTracesToMetrics(td)
	if err != nil {
		c.logger.Error("Failed to convert traces to metrics",
			zap.Error(err),
			zap.Int("input_spans", totalSpans),
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

	c.logger.Debug("Traces converted to metrics",
		zap.Int("input_spans", totalSpans),
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

	c.logger.Debug("Traces successfully processed and metrics sent to next consumer")
	return nil
}

// ConsumeLogs implements connector.Logs.
func (c *profileToMetricsConnector) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	// Log input statistics
	resourceLogsCount := ld.ResourceLogs().Len()
	totalLogRecords := 0
	for i := 0; i < resourceLogsCount; i++ {
		scopeLogs := ld.ResourceLogs().At(i).ScopeLogs()
		for j := 0; j < scopeLogs.Len(); j++ {
			totalLogRecords += scopeLogs.At(j).LogRecords().Len()
		}
	}

	c.logger.Debug("Processing logs",
		zap.Int("resource_logs_count", resourceLogsCount),
		zap.Int("total_log_records", totalLogRecords),
	)

	// Convert logs to metrics using the converter
	metrics, err := c.converter.ConvertLogsToMetrics(ld)
	if err != nil {
		c.logger.Error("Failed to convert logs to metrics",
			zap.Error(err),
			zap.Int("input_log_records", totalLogRecords),
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

	c.logger.Debug("Logs converted to metrics",
		zap.Int("input_log_records", totalLogRecords),
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

	c.logger.Debug("Logs successfully processed and metrics sent to next consumer")
	return nil
}

// ConsumeMetrics implements connector.Metrics.
func (c *profileToMetricsConnector) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	// Log input statistics
	resourceMetricsCount := md.ResourceMetrics().Len()
	totalMetrics := 0
	for i := 0; i < resourceMetricsCount; i++ {
		scopeMetrics := md.ResourceMetrics().At(i).ScopeMetrics()
		for j := 0; j < scopeMetrics.Len(); j++ {
			totalMetrics += scopeMetrics.At(j).Metrics().Len()
		}
	}

	c.logger.Debug("Processing metrics (pass-through)",
		zap.Int("resource_metrics_count", resourceMetricsCount),
		zap.Int("total_metrics", totalMetrics),
	)

	// For metrics input, we can either pass through or transform
	// For now, we'll pass through the metrics
	if err := c.nextConsumer.ConsumeMetrics(ctx, md); err != nil {
		c.logger.Error("Failed to pass through metrics to next consumer",
			zap.Error(err),
			zap.Int("metrics_count", totalMetrics),
		)
		return err
	}

	c.logger.Debug("Metrics successfully passed through to next consumer")
	return nil
}
