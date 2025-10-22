package profiletometrics

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pprofile"
)

// ConverterConfig defines the configuration for the converter
type ConverterConfig struct {
	Metrics       MetricsConfig
	Attributes    []AttributeConfig
	ProcessFilter ProcessFilterConfig
	PatternFilter PatternFilterConfig
	ThreadFilter  ThreadFilterConfig
}

// Converter converts profiling data to metrics
type Converter struct {
	config *ConverterConfig
}

// NewConverter creates a new profile to metrics converter
func NewConverter(cfg *ConverterConfig) (*Converter, error) {
	return &Converter{
		config: cfg,
	}, nil
}

// ConvertProfilesToMetrics converts profiling data to metrics
func (c *Converter) ConvertProfilesToMetrics(ctx context.Context, profiles pprofile.Profiles) (pmetric.Metrics, error) {
	metrics := pmetric.NewMetrics()
	resourceMetrics := metrics.ResourceMetrics().AppendEmpty()

	// Process each profile
	for i := 0; i < profiles.ResourceProfiles().Len(); i++ {
		resourceProfile := profiles.ResourceProfiles().At(i)

		// Extract attributes from resource
		resourceAttributes := c.extractResourceAttributes(resourceProfile.Resource())

		// Process each scope profile
		for j := 0; j < resourceProfile.ScopeProfiles().Len(); j++ {
			scopeProfile := resourceProfile.ScopeProfiles().At(j)

			// Process each profile
			for k := 0; k < scopeProfile.Profiles().Len(); k++ {
				profile := scopeProfile.Profiles().At(k)

				// Extract profile-specific attributes
				profileAttributes := c.extractProfileAttributes(profile, resourceAttributes)

				// Generate metrics based on configuration
				c.generateMetricsFromProfile(profile, profileAttributes, resourceMetrics)
			}
		}
	}

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
func (c *Converter) extractProfileAttributes(profile pprofile.Profile, resourceAttributes map[string]string) map[string]string {
	attributes := make(map[string]string)

	// Copy resource attributes
	for k, v := range resourceAttributes {
		attributes[k] = v
	}

	// Extract attributes based on configuration rules
	for _, attr := range c.config.Attributes {
		value := c.extractAttributeValue(profile, attr)
		if value != "" {
			attributes[attr.Name] = value
		}
	}

	return attributes
}

// extractAttributeValue extracts a single attribute value based on the rule
func (c *Converter) extractAttributeValue(profile pprofile.Profile, attr AttributeConfig) string {
	switch attr.Type {
	case "literal":
		return attr.Value
	case "regex":
		// For now, return the configured value directly
		// In a real implementation, you would apply regex matching
		return attr.Value
	default:
		return attr.Value
	}
}

// generateMetricsFromProfile generates metrics from profile data
func (c *Converter) generateMetricsFromProfile(profile pprofile.Profile, attributes map[string]string, resourceMetrics pmetric.ResourceMetrics) {
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
	// For now, always return true - in a real implementation you would compile and match the pattern
	return true
}

// matchesProcessFilter checks if the profile matches the process filter
func (c *Converter) matchesProcessFilter(attributes map[string]string) bool {
	if !c.config.ProcessFilter.Enabled {
		return true // No filter configured
	}

	_, exists := attributes["process_name"]
	if !exists {
		return false
	}

	// For now, always return true - in a real implementation you would compile and match the pattern
	return true
}

// generateCPUTimeMetrics generates CPU time metrics from profile data
func (c *Converter) generateCPUTimeMetrics(profile pprofile.Profile, attributes map[string]string, scopeMetrics pmetric.ScopeMetrics) {
	metric := scopeMetrics.Metrics().AppendEmpty()
	metric.SetName(c.config.Metrics.CPU.Name)
	metric.SetDescription("CPU time in seconds")

	// Create a gauge metric for CPU time
	gauge := metric.SetEmptyGauge()

	// Calculate CPU time from profile samples
	cpuTime := c.calculateCPUTime(profile)

	dataPoint := gauge.DataPoints().AppendEmpty()
	dataPoint.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dataPoint.SetDoubleValue(cpuTime)

	// Add attributes to the data point
	for key, value := range attributes {
		dataPoint.Attributes().PutStr(key, value)
	}
}

// generateMemoryAllocationMetrics generates memory allocation metrics from profile data
func (c *Converter) generateMemoryAllocationMetrics(profile pprofile.Profile, attributes map[string]string, scopeMetrics pmetric.ScopeMetrics) {
	metric := scopeMetrics.Metrics().AppendEmpty()
	metric.SetName(c.config.Metrics.Memory.Name)
	metric.SetDescription("Memory allocation in bytes")

	// Create a gauge metric for memory allocation
	gauge := metric.SetEmptyGauge()

	// Calculate memory allocation from profile samples
	memoryAllocation := c.calculateMemoryAllocation(profile)

	dataPoint := gauge.DataPoints().AppendEmpty()
	dataPoint.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dataPoint.SetDoubleValue(memoryAllocation)

	// Add attributes to the data point
	for key, value := range attributes {
		dataPoint.Attributes().PutStr(key, value)
	}
}

// calculateCPUTime calculates CPU time from profile samples
func (c *Converter) calculateCPUTime(profile pprofile.Profile) float64 {
	var totalCPUTime float64

	// Sum up CPU time from all samples
	for i := 0; i < profile.Sample().Len(); i++ {
		sample := profile.Sample().At(i)

		// Look for CPU time in sample values
		values := sample.Values()
		for j := 0; j < values.Len(); j++ {
			value := values.At(j)
			// Assuming CPU time is the first value or has a specific type
			// This would need to be adjusted based on actual profile data structure
			totalCPUTime += float64(value)
		}
	}

	return totalCPUTime
}

// calculateMemoryAllocation calculates memory allocation from profile samples
func (c *Converter) calculateMemoryAllocation(profile pprofile.Profile) float64 {
	var totalMemoryAllocation float64

	// Sum up memory allocation from all samples
	for i := 0; i < profile.Sample().Len(); i++ {
		sample := profile.Sample().At(i)

		// Look for memory allocation in sample values
		// This would need to be adjusted based on actual profile data structure
		values := sample.Values()
		for j := 0; j < values.Len(); j++ {
			value := values.At(j)
			// Assuming memory allocation is the second value or has a specific type
			if j == 1 { // Adjust index based on actual data structure
				totalMemoryAllocation += float64(value)
			}
		}
	}

	return totalMemoryAllocation
}
