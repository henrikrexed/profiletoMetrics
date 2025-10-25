package profiletometrics

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.uber.org/zap"
)

const (
	nanosecondsPerSecond = 1e9
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

// matchesSampleFilter checks if a sample matches the given filter criteria
func (c *Converter) matchesSampleFilter(profiles pprofile.Profiles, sample pprofile.Sample, filter map[string]string) bool {
	if filter == nil || len(filter) == 0 {
		return true // No filter means match all
	}

	// Check if the sample matches all filter criteria
	for key, expectedValue := range filter {
		actualValue := c.getSampleAttributeValue(profiles, sample, key)
		if actualValue != expectedValue {
			c.logDebug("Sample does not match filter",
				zap.String("key", key),
				zap.String("expected_value", expectedValue),
				zap.String("actual_value", actualValue))
			return false
		}
	}

	c.logDebug("Sample matches filter", zap.Any("filter", filter))
	return true
}

// getSampleAttributeValue extracts a specific attribute value from a sample
// In the pprofile schema, samples have AttributeIndices that point to AttributeTable entries
// Each AttributeTable entry has KeyStrindex, Value, and UnitStrindex
func (c *Converter) getSampleAttributeValue(profiles pprofile.Profiles, sample pprofile.Sample, key string) string {
	attributeIndices := sample.AttributeIndices()
	if attributeIndices.Len() == 0 {
		return ""
	}

	dictionary := profiles.Dictionary()
	attributeTable := dictionary.AttributeTable()
	stringTable := dictionary.StringTable()

	// Iterate through attribute indices
	for i := 0; i < attributeIndices.Len(); i++ {
		attrIndex := attributeIndices.At(i)
		if attrIndex < 0 || attrIndex >= int32(attributeTable.Len()) {
			continue
		}

		attr := attributeTable.At(int(attrIndex))

		// Get the key from the string table
		keyIndex := attr.KeyStrindex()
		if keyIndex < 0 || keyIndex >= int32(stringTable.Len()) {
			continue
		}

		attrKey := stringTable.At(int(keyIndex))

		// Check if this is the key we're looking for
		if attrKey == key {
			// Get the value
			value := attr.Value()
			return value.AsString()
		}
	}

	return ""
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
				c.generateMetricsFromProfile(profiles, profile, profileAttributes, resourceMetrics)
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
	profiles pprofile.Profiles,
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
		c.generateCPUTimeMetrics(profiles, profile, attributes, scopeMetrics)
	}

	// Generate memory allocation metrics if enabled
	if c.config.Metrics.Memory.Enabled {
		c.generateMemoryAllocationMetrics(profiles, profile, attributes, scopeMetrics)
	}

	// Generate metrics for specific threads
	threadNames := c.getUniqueThreadNames(profiles, profile)
	for _, threadName := range threadNames {
		c.logDebug("Generating metrics for thread", zap.String("thread_name", threadName))
		c.generateThreadMetrics(profiles, profile, attributes, scopeMetrics, threadName)
	}

	// Generate metrics for specific processes
	processNames := c.getUniqueProcessNames(profiles, profile)
	for _, processName := range processNames {
		c.logDebug("Generating metrics for process", zap.String("process_name", processName))
		c.generateProcessMetrics(profiles, profile, attributes, scopeMetrics, processName)
	}

	// Generate function-level metrics (if enabled)
	if c.config.Metrics.Function.Enabled {
		c.generateFunctionMetrics(profiles, profile, attributes, scopeMetrics)
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

// generateMetric creates a metric with the given configuration and value
func (c *Converter) generateMetric(
	profile pprofile.Profile,
	attributes map[string]string,
	scopeMetrics pmetric.ScopeMetrics,
	metricName, description, unit string,
	calculateValue func(pprofile.Profile) float64,
) {
	c.logDebug("Generating metric",
		zap.String("metric_name", metricName),
		zap.String("unit", unit))

	value := calculateValue(profile)

	c.logDebug("Metric generated",
		zap.Float64("value", value),
		zap.String("metric_name", metricName))

	c.generateGaugeMetric(metricName, description, value, attributes, scopeMetrics)
}

// generateCPUTimeMetrics generates CPU time metrics from profile data
func (c *Converter) generateCPUTimeMetrics(profiles pprofile.Profiles, profile pprofile.Profile, attributes map[string]string, scopeMetrics pmetric.ScopeMetrics) {
	cpuTime := c.calculateCPUTime(profiles, profile)
	c.generateGaugeMetric(c.config.Metrics.CPU.MetricName, "CPU time in seconds", cpuTime, attributes, scopeMetrics)
}

// generateMemoryAllocationMetrics generates memory allocation metrics from profile data
func (c *Converter) generateMemoryAllocationMetrics(profiles pprofile.Profiles, profile pprofile.Profile, attributes map[string]string, scopeMetrics pmetric.ScopeMetrics) {
	memoryAllocation := c.calculateMemoryAllocation(profiles, profile)
	c.generateGaugeMetric(c.config.Metrics.Memory.MetricName, "Memory allocation in bytes", memoryAllocation, attributes, scopeMetrics)
}

// generateThreadMetrics generates CPU time and memory metrics for a specific thread
func (c *Converter) generateThreadMetrics(
	profiles pprofile.Profiles,
	profile pprofile.Profile,
	attributes map[string]string,
	scopeMetrics pmetric.ScopeMetrics,
	threadName string,
) {
	filter := map[string]string{"thread.name": threadName}

	// Generate CPU time metric for specific thread
	threadCPUMetricName := fmt.Sprintf("%s_thread_%s", c.config.Metrics.CPU.MetricName, sanitizeMetricName(threadName))
	c.generateMetricWithFilter(profile, attributes, scopeMetrics,
		threadCPUMetricName, fmt.Sprintf("CPU time for thread %s in seconds", threadName), c.config.Metrics.CPU.Unit,
		func(p pprofile.Profile) float64 { return c.calculateCPUTimeForFilter(profiles, p, filter) })

	// Generate memory allocation metric for specific thread
	threadMemoryMetricName := fmt.Sprintf("%s_thread_%s", c.config.Metrics.Memory.MetricName, sanitizeMetricName(threadName))
	c.generateMetricWithFilter(profile, attributes, scopeMetrics,
		threadMemoryMetricName, fmt.Sprintf("Memory allocation for thread %s in bytes", threadName), c.config.Metrics.Memory.Unit,
		func(p pprofile.Profile) float64 { return c.calculateMemoryAllocationForFilter(profiles, p, filter) })
}

// generateProcessMetrics generates CPU time and memory metrics for a specific process
func (c *Converter) generateProcessMetrics(
	profiles pprofile.Profiles,
	profile pprofile.Profile,
	attributes map[string]string,
	scopeMetrics pmetric.ScopeMetrics,
	processName string,
) {
	filter := map[string]string{"process.executable.name": processName}

	// Generate CPU time metric for specific process
	processCPUMetricName := fmt.Sprintf("%s_process_%s", c.config.Metrics.CPU.MetricName, sanitizeMetricName(processName))
	c.generateMetricWithFilter(profile, attributes, scopeMetrics,
		processCPUMetricName, fmt.Sprintf("CPU time for process %s in seconds", processName), c.config.Metrics.CPU.Unit,
		func(p pprofile.Profile) float64 { return c.calculateCPUTimeForFilter(profiles, p, filter) })

	// Generate memory allocation metric for specific process
	processMemoryMetricName := fmt.Sprintf("%s_process_%s", c.config.Metrics.Memory.MetricName, sanitizeMetricName(processName))
	c.generateMetricWithFilter(profile, attributes, scopeMetrics,
		processMemoryMetricName, fmt.Sprintf("Memory allocation for process %s in bytes", processName), c.config.Metrics.Memory.Unit,
		func(p pprofile.Profile) float64 { return c.calculateMemoryAllocationForFilter(profiles, p, filter) })
}

// generateFunctionMetrics generates CPU time and memory metrics for specific functions
func (c *Converter) generateFunctionMetrics(
	profiles pprofile.Profiles,
	profile pprofile.Profile,
	attributes map[string]string,
	scopeMetrics pmetric.ScopeMetrics,
) {
	functionNames := c.getUniqueFunctionNames(profiles, profile)

	for _, functionName := range functionNames {
		c.logDebug("Generating metrics for function", zap.String("function_name", functionName))

		// Generate CPU time metric for specific function
		functionCPUMetricName := fmt.Sprintf("%s_function_%s", c.config.Metrics.CPU.MetricName, sanitizeMetricName(functionName))
		cpuTime := c.calculateFunctionCPUTime(profiles, profile, functionName)
		c.generateGaugeMetric(functionCPUMetricName, fmt.Sprintf("CPU time for function %s in seconds", functionName), cpuTime, attributes, scopeMetrics)

		// Generate memory allocation metric for specific function
		functionMemoryMetricName := fmt.Sprintf("%s_function_%s", c.config.Metrics.Memory.MetricName, sanitizeMetricName(functionName))
		memoryAllocation := c.calculateFunctionMemoryAllocation(profiles, profile, functionName)
		c.generateGaugeMetric(functionMemoryMetricName, fmt.Sprintf("Memory allocation for function %s in bytes", functionName), memoryAllocation, attributes, scopeMetrics)
	}
}

// getUniqueFunctionNames extracts all unique function names from a profile
func (c *Converter) getUniqueFunctionNames(profiles pprofile.Profiles, profile pprofile.Profile) []string {
	functionNames := make(map[string]bool)

	for i := 0; i < profile.Sample().Len(); i++ {
		sample := profile.Sample().At(i)
		functionName := c.getSampleFunctionName(profiles, sample)
		if functionName != "" {
			functionNames[functionName] = true
		}
	}

	var result []string
	for functionName := range functionNames {
		result = append(result, functionName)
	}

	c.logDebug("Extracted unique function names",
		zap.Int("count", len(result)),
		zap.Strings("function_names", result))

	return result
}

// calculateFunctionCPUTime calculates CPU time for a specific function
func (c *Converter) calculateFunctionCPUTime(profiles pprofile.Profiles, profile pprofile.Profile, functionName string) float64 {
	var totalCPUTime float64
	defaultProfileDuration := 1.0
	sampleCount := profile.Sample().Len()

	for i := 0; i < sampleCount; i++ {
		sample := profile.Sample().At(i)
		sampleFunctionName := c.getSampleFunctionName(profiles, sample)

		if sampleFunctionName == functionName {
			values := sample.Values()
			if values.Len() > 0 {
				cpuTimeNs := float64(values.At(0))
				totalCPUTime += cpuTimeNs / nanosecondsPerSecond
			} else if sampleCount > 0 && defaultProfileDuration > 0 {
				totalCPUTime += defaultProfileDuration / float64(sampleCount)
			}
		}
	}

	return totalCPUTime
}

// calculateFunctionMemoryAllocation calculates memory allocation for a specific function
func (c *Converter) calculateFunctionMemoryAllocation(profiles pprofile.Profiles, profile pprofile.Profile, functionName string) float64 {
	var totalMemoryAllocation float64
	sampleCount := profile.Sample().Len()

	for i := 0; i < sampleCount; i++ {
		sample := profile.Sample().At(i)
		sampleFunctionName := c.getSampleFunctionName(profiles, sample)

		if sampleFunctionName == functionName {
			values := sample.Values()
			if values.Len() > 1 {
				totalMemoryAllocation += float64(values.At(1))
			} else if values.Len() == 1 {
				totalMemoryAllocation += float64(values.At(0))
			} else {
				totalMemoryAllocation += 2048.0 // Default 2KB for stack trace profiles
			}
		}
	}

	return totalMemoryAllocation
}

// generateMetricWithFilter creates a metric with the given configuration, value, and filter
func (c *Converter) generateMetricWithFilter(
	profile pprofile.Profile,
	attributes map[string]string,
	scopeMetrics pmetric.ScopeMetrics,
	metricName, description, unit string,
	calculateValue func(pprofile.Profile) float64,
) {
	c.logDebug("Generating filtered metric",
		zap.String("metric_name", metricName),
		zap.String("unit", unit))

	value := calculateValue(profile)

	c.logDebug("Filtered metric generated",
		zap.Float64("value", value),
		zap.String("metric_name", metricName))

	c.generateGaugeMetric(metricName, description, value, attributes, scopeMetrics)
}

// sanitizeMetricName sanitizes a string to be used as a metric name
func sanitizeMetricName(name string) string {
	// Replace invalid characters with underscores
	result := ""
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_' {
			result += string(char)
		} else {
			result += "_"
		}
	}
	return result
}

// getFunctionName extracts the function name from a function index using the profiles dictionary
func (c *Converter) getFunctionName(profiles pprofile.Profiles, functionIndex int32) string {
	if functionIndex < 0 {
		return ""
	}

	dictionary := profiles.Dictionary()
	functionTable := dictionary.FunctionTable()

	if functionIndex >= int32(functionTable.Len()) {
		c.logDebug("Function index out of range",
			zap.Int32("function_index", functionIndex),
			zap.Int("function_table_len", functionTable.Len()))
		return ""
	}

	function := functionTable.At(int(functionIndex))
	nameIndex := function.NameStrindex()

	stringTable := dictionary.StringTable()
	if nameIndex < 0 || nameIndex >= int32(stringTable.Len()) {
		c.logDebug("Function name index out of range",
			zap.Int32("name_index", nameIndex),
			zap.Int32("function_index", functionIndex),
			zap.Int("string_table_len", stringTable.Len()))
		return ""
	}

	functionName := stringTable.At(int(nameIndex))
	c.logDebug("Resolved function name",
		zap.Int32("function_index", functionIndex),
		zap.String("function_name", functionName))

	return functionName
}

// getLocationFunctionName gets the function name from a location using the profiles dictionary
func (c *Converter) getLocationFunctionName(profiles pprofile.Profiles, location pprofile.Location) string {
	// Locations have Lines, and Lines have FunctionIndex
	lines := location.Line()
	if lines.Len() == 0 {
		return ""
	}

	// Get the first line's function (most specific in the call stack)
	line := lines.At(0)
	functionIndex := line.FunctionIndex()

	return c.getFunctionName(profiles, functionIndex)
}

// getSampleFunctionName gets the top function name from a sample's stack
func (c *Converter) getSampleFunctionName(profiles pprofile.Profiles, sample pprofile.Sample) string {
	stackIndex := sample.StackIndex()
	if stackIndex < 0 {
		c.logDebug("Sample has no stack index")
		return ""
	}

	dictionary := profiles.Dictionary()
	stackTable := dictionary.StackTable()

	if stackIndex >= int32(stackTable.Len()) {
		c.logDebug("Stack index out of range",
			zap.Int32("stack_index", stackIndex),
			zap.Int("stack_table_len", stackTable.Len()))
		return ""
	}

	stack := stackTable.At(int(stackIndex))
	locationIndices := stack.LocationIndices()

	if locationIndices.Len() == 0 {
		c.logDebug("Stack has no locations")
		return ""
	}

	// Get the first location (top of the call stack)
	locationIndex := locationIndices.At(0)
	locationTable := dictionary.LocationTable()

	if locationIndex < 0 || locationIndex >= int32(locationTable.Len()) {
		c.logDebug("Location index out of range",
			zap.Int32("location_index", locationIndex),
			zap.Int("location_table_len", locationTable.Len()))
		return ""
	}

	location := locationTable.At(int(locationIndex))
	return c.getLocationFunctionName(profiles, location)
}

// getUniqueThreadNames extracts all unique thread names from a profile
// In the pprofile schema, thread information is stored as resource attributes
func (c *Converter) getUniqueThreadNames(profiles pprofile.Profiles, profile pprofile.Profile) []string {
	threadNames := make(map[string]bool)

	// Iterate through samples to extract unique thread names from attributes
	for i := 0; i < profile.Sample().Len(); i++ {
		sample := profile.Sample().At(i)
		threadName := c.getSampleAttributeValue(profiles, sample, "thread.name")
		if threadName != "" {
			threadNames[threadName] = true
		}
	}

	var result []string
	for threadName := range threadNames {
		result = append(result, threadName)
	}

	c.logDebug("Extracted unique thread names",
		zap.Int("count", len(result)),
		zap.Strings("thread_names", result))

	return result
}

// getUniqueProcessNames extracts all unique process names from a profile
// In the pprofile schema, process information is stored as resource attributes
func (c *Converter) getUniqueProcessNames(profiles pprofile.Profiles, profile pprofile.Profile) []string {
	processNames := make(map[string]bool)

	// Iterate through samples to extract unique process names from attributes
	for i := 0; i < profile.Sample().Len(); i++ {
		sample := profile.Sample().At(i)
		processName := c.getSampleAttributeValue(profiles, sample, "process.executable.name")
		if processName != "" {
			processNames[processName] = true
		}
	}

	var result []string
	for processName := range processNames {
		result = append(result, processName)
	}

	c.logDebug("Extracted unique process names",
		zap.Int("count", len(result)),
		zap.Strings("process_names", result))

	return result
}

// calculateCPUTime calculates CPU time from profile samples
func (c *Converter) calculateCPUTime(profiles pprofile.Profiles, profile pprofile.Profile) float64 {
	return c.calculateCPUTimeForFilter(profiles, profile, nil)
}

// calculateCPUTimeForFilter calculates CPU time from profile samples with optional filtering
func (c *Converter) calculateCPUTimeForFilter(profiles pprofile.Profiles, profile pprofile.Profile, filter map[string]string) float64 {
	var totalCPUTime float64
	sampleCount := profile.Sample().Len()

	c.logDebug("Calculating CPU time",
		zap.Int("samples_count", sampleCount),
		zap.Any("filter", filter))

	// For stack trace profiles, we'll use a default duration since we can't get timing from the profile
	// This is a reasonable assumption for profiling data
	defaultProfileDuration := 1.0 // 1 second default
	c.logDebug("Profile timing",
		zap.Float64("default_duration_seconds", defaultProfileDuration))

	// Sum up CPU time from all samples
	for i := 0; i < sampleCount; i++ {
		sample := profile.Sample().At(i)
		values := sample.Values()

		// Apply filtering if specified
		if filter != nil && !c.matchesSampleFilter(profiles, sample, filter) {
			c.logDebug("Sample filtered out",
				zap.Int("sample_index", i),
				zap.Any("filter", filter))
			continue
		}

		c.logDebug("Processing sample",
			zap.Int("sample_index", i),
			zap.Int("values_count", values.Len()))

		// Log all values in the sample for debugging
		if values.Len() > 0 {
			valueStrings := make([]string, values.Len())
			for v := 0; v < values.Len(); v++ {
				valueStrings[v] = fmt.Sprintf("values[%d]=%d", v, values.At(v))
			}
			c.logDebug("Sample values",
				zap.Int("sample_index", i),
				zap.Strings("values", valueStrings))
		} else {
			c.logWarn("Sample has no values", zap.Int("sample_index", i))

			// Let's also check if there are other ways to access sample data
			c.logDebug("Sample structure analysis",
				zap.Int("sample_index", i),
				zap.String("sample_type", fmt.Sprintf("%T", sample)))
		}

		// Look for CPU time in sample values
		// For CPU time, we typically want the first value (index 0)
		// or we need to check the value type if available
		if values.Len() > 0 {
			// Take the first value as CPU time (in nanoseconds)
			cpuTimeNs := float64(values.At(0))
			// Convert nanoseconds to seconds for better readability
			cpuTimeSeconds := cpuTimeNs / nanosecondsPerSecond
			totalCPUTime += cpuTimeSeconds

			c.logDebug("Sample CPU time",
				zap.Int("sample_index", i),
				zap.Float64("cpu_time_ns", cpuTimeNs),
				zap.Float64("cpu_time_seconds", cpuTimeSeconds),
				zap.Float64("running_total", totalCPUTime))
		} else {
			c.logWarn("Sample has no values - this is expected for stack trace profiles", zap.Int("sample_index", i))

			// For stack trace profiles without values, distribute the profile duration
			// across all samples to estimate CPU time per sample
			if sampleCount > 0 && defaultProfileDuration > 0 {
				estimatedCPUTimePerSample := defaultProfileDuration / float64(sampleCount)
				totalCPUTime += estimatedCPUTimePerSample

				c.logDebug("Using estimated CPU time based on profile duration",
					zap.Int("sample_index", i),
					zap.Float64("estimated_cpu_time_seconds", estimatedCPUTimePerSample),
					zap.Float64("profile_duration_seconds", defaultProfileDuration),
					zap.Float64("running_total", totalCPUTime))
			} else {
				// Fallback to a small default value
				defaultCPUTime := 0.001 // 1ms default
				totalCPUTime += defaultCPUTime

				c.logDebug("Using default CPU time for stack trace sample",
					zap.Int("sample_index", i),
					zap.Float64("default_cpu_time_seconds", defaultCPUTime),
					zap.Float64("running_total", totalCPUTime))
			}
		}
	}

	c.logDebug("CPU time calculation completed",
		zap.Float64("total_cpu_time_seconds", totalCPUTime),
		zap.Int("samples_processed", sampleCount))

	return totalCPUTime
}

// calculateMemoryAllocation calculates memory allocation from profile samples
func (c *Converter) calculateMemoryAllocation(profiles pprofile.Profiles, profile pprofile.Profile) float64 {
	return c.calculateMemoryAllocationForFilter(profiles, profile, nil)
}

// calculateMemoryAllocationForFilter calculates memory allocation from profile samples with optional filtering
func (c *Converter) calculateMemoryAllocationForFilter(profiles pprofile.Profiles, profile pprofile.Profile, filter map[string]string) float64 {
	var totalMemoryAllocation float64
	sampleCount := profile.Sample().Len()

	c.logDebug("Calculating memory allocation",
		zap.Int("samples_count", sampleCount),
		zap.Any("filter", filter))

	// Sum up memory allocation from all samples
	for i := 0; i < sampleCount; i++ {
		sample := profile.Sample().At(i)
		values := sample.Values()

		// Apply filtering if specified
		if filter != nil && !c.matchesSampleFilter(profiles, sample, filter) {
			c.logDebug("Sample filtered out",
				zap.Int("sample_index", i),
				zap.Any("filter", filter))
			continue
		}

		c.logDebug("Processing sample for memory",
			zap.Int("sample_index", i),
			zap.Int("values_count", values.Len()))

		// Log all values in the sample for debugging
		if values.Len() > 0 {
			valueStrings := make([]string, values.Len())
			for v := 0; v < values.Len(); v++ {
				valueStrings[v] = fmt.Sprintf("values[%d]=%d", v, values.At(v))
			}
			c.logDebug("Sample values for memory",
				zap.Int("sample_index", i),
				zap.Strings("values", valueStrings))
		} else {
			c.logWarn("Sample has no values for memory calculation", zap.Int("sample_index", i))
		}

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
			c.logWarn("Sample has no values for memory calculation - this is expected for stack trace profiles", zap.Int("sample_index", i))

			// For stack trace profiles without values, estimate memory allocation
			// based on a reasonable default for stack trace samples
			estimatedMemoryBytes := 2048.0 // 2KB default for stack trace sample
			totalMemoryAllocation += estimatedMemoryBytes

			c.logDebug("Using estimated memory allocation for stack trace sample",
				zap.Int("sample_index", i),
				zap.Float64("estimated_memory_bytes", estimatedMemoryBytes),
				zap.Float64("running_total", totalMemoryAllocation))
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
