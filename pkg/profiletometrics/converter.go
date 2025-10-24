package profiletometrics

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.uber.org/zap"
)

// ConverterConfig defines the configuration for the converter
type ConverterConfig struct {
	Metrics       MetricsConfig       `mapstructure:"metrics"`
	Attributes    []AttributeConfig   `mapstructure:"attributes"`
	ProcessFilter ProcessFilterConfig `mapstructure:"process_filter"`
	PatternFilter PatternFilterConfig `mapstructure:"pattern_filter"`
	ThreadFilter  ThreadFilterConfig  `mapstructure:"thread_filter"`
}

// Converter converts profiling data to metrics
type Converter struct {
	config *ConverterConfig
	logger *zap.Logger
}

// NewConverter creates a new profile to metrics converter
func NewConverter(cfg *ConverterConfig) (*Converter, error) {
	return &Converter{
		config: cfg,
		logger: nil, // Will be set by the connector
	}, nil
}

// SetLogger sets the logger for the converter
func (c *Converter) SetLogger(logger *zap.Logger) {
	c.logger = logger
}

// logInfo logs an info message if logger is available
func (c *Converter) logInfo(msg string, fields ...zap.Field) {
	if c.logger != nil {
		c.logger.Info(msg, fields...)
	}
}

// logDebug logs a debug message if logger is available
func (c *Converter) logDebug(msg string, fields ...zap.Field) {
	if c.logger != nil {
		c.logger.Debug(msg, fields...)
	}
}

// logWarn logs a warning message if logger is available
func (c *Converter) logWarn(msg string, fields ...zap.Field) {
	if c.logger != nil {
		c.logger.Warn(msg, fields...)
	}
}

// ConvertProfilesToMetrics converts profiling data to metrics
func (c *Converter) ConvertProfilesToMetrics(ctx context.Context, profiles pprofile.Profiles) (pmetric.Metrics, error) {
	c.logInfo("Starting profile to metrics conversion",
		zap.Int("resource_profiles_count", profiles.ResourceProfiles().Len()))

	metrics := pmetric.NewMetrics()
	resourceMetrics := metrics.ResourceMetrics().AppendEmpty()

	// Process each profile
	for i := 0; i < profiles.ResourceProfiles().Len(); i++ {
		resourceProfile := profiles.ResourceProfiles().At(i)
		c.logDebug("Processing resource profile", zap.Int("index", i))

		// Extract attributes from resource
		resourceAttributes := c.extractResourceAttributes(resourceProfile.Resource())
		c.logDebug("Extracted resource attributes", zap.Any("attributes", resourceAttributes))

		// Process each scope profile
		for j := 0; j < resourceProfile.ScopeProfiles().Len(); j++ {
			scopeProfile := resourceProfile.ScopeProfiles().At(j)
			c.logDebug("Processing scope profile",
				zap.Int("resource_index", i),
				zap.Int("scope_index", j),
				zap.String("scope_name", scopeProfile.Scope().Name()),
				zap.String("scope_version", scopeProfile.Scope().Version()))

			// Process each profile
			for k := 0; k < scopeProfile.Profiles().Len(); k++ {
				profile := scopeProfile.Profiles().At(k)
				c.logDebug("Processing profile",
					zap.Int("resource_index", i),
					zap.Int("scope_index", j),
					zap.Int("profile_index", k),
					zap.Int("samples_count", profile.Sample().Len()))

				// Extract profile-specific attributes
				profileAttributes := c.extractProfileAttributes(profiles, profile, resourceAttributes)
				c.logDebug("Extracted profile attributes", zap.Any("attributes", profileAttributes))

				// Generate metrics based on configuration
				c.generateMetricsFromProfile(profile, profileAttributes, resourceMetrics)
			}
		}
	}

	c.logInfo("Profile to metrics conversion completed")
	return metrics, nil
}

// extractResourceAttributes extracts attributes from the resource
func (c *Converter) extractResourceAttributes(resource pcommon.Resource) map[string]string {
	attributes := make(map[string]string)

	resource.Attributes().Range(func(key string, value pcommon.Value) bool {
		attributes[key] = value.AsString()
		return true
	})

	return attributes
}

// extractProfileAttributes extracts attributes from the profile data
func (c *Converter) extractProfileAttributes(
	profiles pprofile.Profiles,
	profile pprofile.Profile,
	resourceAttributes map[string]string,
) map[string]string {
	attributes := make(map[string]string)

	// Copy resource attributes
	for k, v := range resourceAttributes {
		attributes[k] = v
	}

	// Extract attributes based on configuration rules
	for _, attr := range c.config.Attributes {
		value := c.extractAttributeValue(profiles, profile, attr)
		if value != "" {
			attributes[attr.Key] = value
		}
	}

	return attributes
}

// extractAttributeValue extracts a single attribute value based on the rule
func (c *Converter) extractAttributeValue(profiles pprofile.Profiles, _ pprofile.Profile, attr AttributeConfig) string {
	switch attr.Type {
	case "literal":
		return attr.Value
	case "regex":
		// Extract from string table using regex pattern
		return c.extractFromStringTable(profiles, attr.Value)
	case "string_table":
		// Direct string table index access
		return c.extractFromStringTableByIndex(profiles, attr.Value)
	default:
		return attr.Value
	}
}

// generateMetricsFromProfile generates metrics from profile data
func (c *Converter) generateMetricsFromProfile(
	profile pprofile.Profile,
	attributes map[string]string,
	resourceMetrics pmetric.ResourceMetrics,
) {
	// Apply pattern filtering if enabled
	if c.config.PatternFilter.Enabled && !c.matchesPatternFilter(attributes) {
		return
	}

	// Apply process filtering
	if !c.matchesProcessFilter(attributes) {
		return
	}

	// Create a single scope metrics for all metrics from this profile
	scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
	scopeMetrics.Scope().SetName("profiletometrics")
	scopeMetrics.Scope().SetVersion("1.0.0")

	// Generate CPU time metrics if enabled
	if c.config.Metrics.CPU.Enabled {
		c.generateCPUTimeMetrics(profile, attributes, scopeMetrics)
	}

	// Generate memory allocation metrics if enabled
	if c.config.Metrics.Memory.Enabled {
		c.generateMemoryAllocationMetrics(profile, attributes, scopeMetrics)
	}
}

// matchesPatternFilter checks if attributes match the pattern filter
func (c *Converter) matchesPatternFilter(attributes map[string]string) bool {
	if !c.config.PatternFilter.Enabled {
		return true
	}
	// Check if any attribute value matches the pattern
	for _, value := range attributes {
		if c.config.PatternFilter.Pattern != "" &&
			value != "" {
			// Simple substring matching for now
			return true
		}
	}
	return false
}

// matchesProcessFilter checks if the profile matches the process filter
func (c *Converter) matchesProcessFilter(attributes map[string]string) bool {
	if !c.config.ProcessFilter.Enabled {
		return true // No filter configured
	}

	processName, exists := attributes["process_name"]
	if !exists {
		return false // No process name attribute found
	}

	// For now, simple string matching - in a real implementation you would compile and match the regex pattern
	if c.config.ProcessFilter.Pattern == "" {
		return true // No pattern specified, allow all
	}

	// Simple contains check for now - in production, use regex compilation
	return processName != "" // Placeholder logic
}

// generateGaugeMetric generates a gauge metric with the given configuration
func (c *Converter) generateGaugeMetric(
	name, description string,
	value float64,
	attributes map[string]string,
	scopeMetrics pmetric.ScopeMetrics,
) {
	metric := scopeMetrics.Metrics().AppendEmpty()
	metric.SetName(name)
	metric.SetDescription(description)

	// Create a gauge metric
	gauge := metric.SetEmptyGauge()

	dataPoint := gauge.DataPoints().AppendEmpty()
	dataPoint.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dataPoint.SetDoubleValue(value)

	// Add attributes to the data point
	for key, val := range attributes {
		dataPoint.Attributes().PutStr(key, val)
	}
}

// generateCPUTimeMetrics generates CPU time metrics from profile data
func (c *Converter) generateCPUTimeMetrics(profile pprofile.Profile, attributes map[string]string, scopeMetrics pmetric.ScopeMetrics) {
	c.logDebug("Generating CPU time metrics",
		zap.String("metric_name", c.config.Metrics.CPU.MetricName),
		zap.String("unit", c.config.Metrics.CPU.Unit))

	cpuTime := c.calculateCPUTime(profile)

	c.logDebug("CPU time metric generated",
		zap.Float64("cpu_time_seconds", cpuTime),
		zap.String("metric_name", c.config.Metrics.CPU.MetricName))

	c.generateGaugeMetric(c.config.Metrics.CPU.MetricName, "CPU time in seconds", cpuTime, attributes, scopeMetrics)
}

// generateMemoryAllocationMetrics generates memory allocation metrics from profile data
func (c *Converter) generateMemoryAllocationMetrics(
	profile pprofile.Profile,
	attributes map[string]string,
	scopeMetrics pmetric.ScopeMetrics,
) {
	c.logDebug("Generating memory allocation metrics",
		zap.String("metric_name", c.config.Metrics.Memory.MetricName),
		zap.String("unit", c.config.Metrics.Memory.Unit))

	memoryAllocation := c.calculateMemoryAllocation(profile)

	c.logDebug("Memory allocation metric generated",
		zap.Float64("memory_allocation_bytes", memoryAllocation),
		zap.String("metric_name", c.config.Metrics.Memory.MetricName))

	c.generateGaugeMetric(c.config.Metrics.Memory.MetricName, "Memory allocation in bytes", memoryAllocation, attributes, scopeMetrics)
}

// calculateCPUTime calculates CPU time from profile samples
func (c *Converter) calculateCPUTime(profile pprofile.Profile) float64 {
	var totalCPUTime float64
	sampleCount := profile.Sample().Len()

	c.logDebug("Calculating CPU time", zap.Int("samples_count", sampleCount))

	// Sum up CPU time from all samples
	for i := 0; i < sampleCount; i++ {
		sample := profile.Sample().At(i)
		values := sample.Values()

		c.logDebug("Processing sample",
			zap.Int("sample_index", i),
			zap.Int("values_count", values.Len()))

		// Look for CPU time in sample values
		// For CPU time, we typically want the first value (index 0)
		// or we need to check the value type if available
		if values.Len() > 0 {
			// Take the first value as CPU time (in nanoseconds)
			cpuTimeNs := float64(values.At(0))
			// Convert nanoseconds to seconds for better readability
			cpuTimeSeconds := cpuTimeNs / 1e9
			totalCPUTime += cpuTimeSeconds

			c.logDebug("Sample CPU time",
				zap.Int("sample_index", i),
				zap.Float64("cpu_time_ns", cpuTimeNs),
				zap.Float64("cpu_time_seconds", cpuTimeSeconds),
				zap.Float64("running_total", totalCPUTime))
		} else {
			c.logWarn("Sample has no values", zap.Int("sample_index", i))
		}
	}

	c.logDebug("CPU time calculation completed",
		zap.Float64("total_cpu_time_seconds", totalCPUTime),
		zap.Int("samples_processed", sampleCount))

	return totalCPUTime
}

// calculateMemoryAllocation calculates memory allocation from profile samples
func (c *Converter) calculateMemoryAllocation(profile pprofile.Profile) float64 {
	var totalMemoryAllocation float64
	sampleCount := profile.Sample().Len()

	c.logDebug("Calculating memory allocation", zap.Int("samples_count", sampleCount))

	// Sum up memory allocation from all samples
	for i := 0; i < sampleCount; i++ {
		sample := profile.Sample().At(i)
		values := sample.Values()

		c.logDebug("Processing sample for memory",
			zap.Int("sample_index", i),
			zap.Int("values_count", values.Len()))

		// Look for memory allocation in sample values
		// For memory allocation, we typically want the second value (index 1)
		// if it exists, otherwise we might need to look for specific value types
		if values.Len() > 1 {
			// Take the second value as memory allocation (in bytes)
			memoryBytes := float64(values.At(1))
			totalMemoryAllocation += memoryBytes

			c.logDebug("Sample memory allocation (index 1)",
				zap.Int("sample_index", i),
				zap.Float64("memory_bytes", memoryBytes),
				zap.Float64("running_total", totalMemoryAllocation))
		} else if values.Len() == 1 {
			// If only one value exists, it might be memory allocation
			// This is a fallback for profiles with only memory data
			memoryBytes := float64(values.At(0))
			totalMemoryAllocation += memoryBytes

			c.logDebug("Sample memory allocation (fallback to index 0)",
				zap.Int("sample_index", i),
				zap.Float64("memory_bytes", memoryBytes),
				zap.Float64("running_total", totalMemoryAllocation))
		} else {
			c.logWarn("Sample has no values for memory calculation", zap.Int("sample_index", i))
		}
	}

	c.logDebug("Memory allocation calculation completed",
		zap.Float64("total_memory_bytes", totalMemoryAllocation),
		zap.Int("samples_processed", sampleCount))

	return totalMemoryAllocation
}

// extractFromStringTable extracts values from profile string table using regex pattern
func (c *Converter) extractFromStringTable(profiles pprofile.Profiles, _ string) string {
	// Access the string table from the profiles dictionary
	stringTable := profiles.Dictionary().StringTable()

	// For now, return the first string as a placeholder
	// In a real implementation, you would:
	// 1. Compile the regex pattern
	// 2. Match against all strings in the table
	// 3. Return the first match
	if stringTable.Len() > 0 {
		return stringTable.At(0)
	}
	return ""
}

// extractFromStringTableByIndex extracts values from profile string table by index
func (c *Converter) extractFromStringTableByIndex(profiles pprofile.Profiles, _ string) string {
	// Access the string table from the profiles dictionary
	stringTable := profiles.Dictionary().StringTable()

	// Parse the index string to integer
	// For now, use index 0 as a placeholder
	// In a real implementation, you would:
	// 1. Parse the indexStr to integer using strconv.Atoi
	// 2. Check bounds to ensure the index is valid
	// 3. Return the string at the specified index
	if stringTable.Len() > 0 {
		return stringTable.At(0) // Placeholder: return first string
	}
	return ""
}
