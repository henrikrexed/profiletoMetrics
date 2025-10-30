package profiletometrics

import (
	"context"
	"fmt"
	"time"

	"regexp"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.uber.org/zap"
)

const (
	nanosecondsPerSecond = 1e9

	// Attribute extraction types
	attrTypeLiteral     = "literal"
	attrTypeRegex       = "regex"
	attrTypeStringTable = "string_table"
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
	if len(filter) == 0 {
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
	return getSampleAttributeValueCommon(profiles, sample, key)
}

// ConvertProfilesToMetrics converts profiling data to metrics
func (c *Converter) ConvertProfilesToMetrics(ctx context.Context, profiles pprofile.Profiles) (pmetric.Metrics, error) {
	c.logInfo("Starting profile to metrics conversion",
		zap.Int("resource_profiles_count", profiles.ResourceProfiles().Len()))

	metrics := pmetric.NewMetrics()
	resourceMetrics := metrics.ResourceMetrics().AppendEmpty()

	iterateProfilesCommon(
		profiles,
		c.extractResourceAttributes,
		func(resourceIndex, scopeIndex, profileIndex int, profile pprofile.Profile, resourceAttributes map[string]string) {
			c.logDebug("Processing profile",
				zap.Int("resource_index", resourceIndex),
				zap.Int("scope_index", scopeIndex),
				zap.Int("profile_index", profileIndex),
				zap.Int("samples_count", profile.Sample().Len()))

			profileAttributes := c.extractProfileAttributes(profiles, profile, resourceAttributes)
			c.logDebug("Extracted profile attributes", zap.Any("attributes", profileAttributes))

			c.generateMetricsFromProfile(profiles, profile, profileAttributes, resourceMetrics)
		},
	)

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
	case attrTypeLiteral:
		return attr.Value
	case attrTypeRegex:
		// Extract from string table using regex pattern
		return c.extractFromStringTable(profiles, attr.Value)
	case attrTypeStringTable:
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
	// pattern_filter deprecated: no-op

	// Apply process filtering against profile samples (process.executable.name), supporting multiple patterns
	// Also, when enabled, restrict metrics generation to matched processes only.
	var matchedProcessNames []string
	if c.config.ProcessFilter.Enabled {
		if !c.profileMatchesProcessFilter(profiles, profile) {
			return
		}
		// Build regexes and filter the discovered processes
		allProcessNames := c.getUniqueProcessNames(profiles, profile)
		var patterns []string
		if len(c.config.ProcessFilter.Patterns) > 0 {
			patterns = c.config.ProcessFilter.Patterns
		} else if c.config.ProcessFilter.Pattern != "" {
			patterns = []string{c.config.ProcessFilter.Pattern}
		}
		regexes := make([]*regexp.Regexp, 0, len(patterns))
		for _, p := range patterns {
			re, err := regexp.Compile(p)
			if err != nil {
				c.logWarn("Invalid process filter pattern - ignoring", zap.String("pattern", p), zap.Error(err))
				continue
			}
			regexes = append(regexes, re)
		}
		for _, name := range allProcessNames {
			for _, re := range regexes {
				if re.MatchString(name) {
					matchedProcessNames = append(matchedProcessNames, name)
					break
				}
			}
		}
		c.logDebug("Process filter matched processes", zap.Strings("process_names", matchedProcessNames))
		if len(matchedProcessNames) == 0 {
			// No processes matched; nothing to emit
			return
		}
	}

	// Create a single scope metrics for all metrics from this profile
	scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
	scopeMetrics.Scope().SetName("profiletometrics")
	scopeMetrics.Scope().SetVersion("1.0.0")

	// If process filter is enabled, skip unfiltered/global metrics; emit only per-process metrics
	if !c.config.ProcessFilter.Enabled {
		// Generate CPU time metrics if enabled
		if c.config.Metrics.CPU.Enabled {
			c.generateCPUTimeMetrics(profiles, profile, attributes, scopeMetrics)
		}
		// Generate memory allocation metrics if enabled
		if c.config.Metrics.Memory.Enabled {
			c.generateMemoryAllocationMetrics(profiles, profile, attributes, scopeMetrics)
		}
	} else {
		c.logDebug("Process filter enabled - skipping global metrics in favor of per-process metrics")
	}

	// Generate metrics for specific processes
	processNames := matchedProcessNames
	if !c.config.ProcessFilter.Enabled {
		processNames = c.getUniqueProcessNames(profiles, profile)
	}
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
	// Backward-compat for existing unit tests: if enabled and no process_name attribute, return false
	if !c.config.ProcessFilter.Enabled {
		return true
	}
	if _, exists := attributes["process_name"]; !exists {
		return false
	}
	return true
}

// profileMatchesProcessFilter checks if the profile contains any process that matches configured patterns
func (c *Converter) profileMatchesProcessFilter(profiles pprofile.Profiles, profile pprofile.Profile) bool {
	if !c.config.ProcessFilter.Enabled {
		return true
	}

	// Build pattern list (prefer list; fallback to single)
	var patterns []string
	if len(c.config.ProcessFilter.Patterns) > 0 {
		patterns = c.config.ProcessFilter.Patterns
	} else if c.config.ProcessFilter.Pattern != "" {
		patterns = []string{c.config.ProcessFilter.Pattern}
	} else {
		return true // enabled but no patterns => allow all
	}

	// Precompile regexes
	regexes := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			c.logWarn("Invalid process filter pattern - ignoring", zap.String("pattern", p), zap.Error(err))
			continue
		}
		regexes = append(regexes, re)
	}
	if len(regexes) == 0 {
		return true // no valid patterns
	}

	// Check unique process names from samples
	processNames := c.getUniqueProcessNames(profiles, profile)
	for _, name := range processNames {
		for _, re := range regexes {
			if re.MatchString(name) {
				c.logDebug("Process filter matched", zap.String("process", name), zap.Strings("patterns", patterns))
				return true
			}
		}
	}

	c.logDebug("Process filter did not match any process", zap.Strings("processes", processNames), zap.Strings("patterns", patterns))
	return false
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
func (c *Converter) generateCPUTimeMetrics(
	profiles pprofile.Profiles,
	profile pprofile.Profile,
	attributes map[string]string,
	scopeMetrics pmetric.ScopeMetrics,
) {
	cpuTime := c.calculateCPUTime(profiles, profile)
	c.generateGaugeMetric(c.config.Metrics.CPU.MetricName, "CPU time in seconds", cpuTime, attributes, scopeMetrics)
}

// generateMemoryAllocationMetrics generates memory allocation metrics from profile data
func (c *Converter) generateMemoryAllocationMetrics(
	profiles pprofile.Profiles,
	profile pprofile.Profile,
	attributes map[string]string,
	scopeMetrics pmetric.ScopeMetrics,
) {
	memoryAllocation := c.calculateMemoryAllocation(profiles, profile)
	c.generateGaugeMetric(c.config.Metrics.Memory.MetricName, "Memory allocation in bytes", memoryAllocation, attributes, scopeMetrics)
}

// generateThreadMetrics generates CPU time and memory metrics for threads with thread.name as attribute
func (c *Converter) generateThreadMetrics(
	profiles pprofile.Profiles,
	profile pprofile.Profile,
	attributes map[string]string,
	scopeMetrics pmetric.ScopeMetrics,
	threadName string,
) {
	c.generateEntityMetrics(profiles, profile, attributes, scopeMetrics, "thread.name", "thread.name", threadName)
}

// generateProcessMetrics generates CPU time and memory metrics for processes with process.name as attribute
func (c *Converter) generateProcessMetrics(
	profiles pprofile.Profiles,
	profile pprofile.Profile,
	attributes map[string]string,
	scopeMetrics pmetric.ScopeMetrics,
	processName string,
) {
	c.generateEntityMetrics(profiles, profile, attributes, scopeMetrics, "process.executable.name", "process.name", processName)
}

// generateEntityMetrics is a generic helper used by thread and process metrics generators
func (c *Converter) generateEntityMetrics(
	profiles pprofile.Profiles,
	profile pprofile.Profile,
	baseAttributes map[string]string,
	scopeMetrics pmetric.ScopeMetrics,
	filterKey string,
	attributeName string,
	attributeValue string,
) {
	filter := map[string]string{filterKey: attributeValue}

	attrs := make(map[string]string)
	for k, v := range baseAttributes {
		attrs[k] = v
	}
	attrs[attributeName] = attributeValue

	cpuTime := c.calculateCPUTimeForFilter(profiles, profile, filter)
	c.generateGaugeMetric(c.config.Metrics.CPU.MetricName, "CPU time in seconds", cpuTime, attrs, scopeMetrics)

	memoryAllocation := c.calculateMemoryAllocationForFilter(profiles, profile, filter)
	c.generateGaugeMetric(c.config.Metrics.Memory.MetricName, "Memory allocation in bytes", memoryAllocation, attrs, scopeMetrics)
}

// generateFunctionMetrics generates CPU time and memory metrics for specific functions
func (c *Converter) generateFunctionMetrics(
	profiles pprofile.Profiles,
	profile pprofile.Profile,
	attributes map[string]string,
	scopeMetrics pmetric.ScopeMetrics,
) {
	c.logDebug("generateFunctionMetrics called - starting function metric generation")

	// Get all function names
	functionNames := c.getUniqueFunctionNames(profiles, profile)

	if len(functionNames) == 0 {
		c.logDebug("No functions found in profile")
		return
	}

	c.logDebug("Generating function-level metrics",
		zap.Int("function_count", len(functionNames)),
		zap.Strings("function_names", functionNames))

	// Precompute function -> filename mapping
	functionToFilename := c.getFunctionFilenameMap(profiles, profile)

	// Create a metric for CPU time with function attributes
	cpuMetricName := c.config.Metrics.CPU.MetricName
	description := "CPU time in seconds"

	cpuMetric := scopeMetrics.Metrics().AppendEmpty()
	cpuMetric.SetName(cpuMetricName)
	cpuMetric.SetDescription(description)
	cpuGauge := cpuMetric.SetEmptyGauge()

	// Create a metric for memory allocation with function attributes
	memoryMetricName := c.config.Metrics.Memory.MetricName
	memDescription := "Memory allocation in bytes"

	memoryMetric := scopeMetrics.Metrics().AppendEmpty()
	memoryMetric.SetName(memoryMetricName)
	memoryMetric.SetDescription(memDescription)
	memoryGauge := memoryMetric.SetEmptyGauge()

	// Get all unique process names to combine with function names
	processNames := c.getUniqueProcessNames(profiles, profile)

	// Create data points for each (process, function) combination
	for _, processName := range processNames {
		for _, functionName := range functionNames {
			c.logDebug("Adding data point for process and function",
				zap.String("process_name", processName),
				zap.String("function_name", functionName))

			// Calculate values for this process and function combination
			cpuTime := c.calculateFunctionCPUTimeForProcess(profiles, profile, processName, functionName)
			memoryAllocation := c.calculateFunctionMemoryAllocationForProcess(profiles, profile, processName, functionName)

			// Create CPU data point with both process and function attributes
			cpuDataPoint := cpuGauge.DataPoints().AppendEmpty()
			cpuDataPoint.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
			cpuDataPoint.SetDoubleValue(cpuTime)

			// Add base attributes
			for key, val := range attributes {
				cpuDataPoint.Attributes().PutStr(key, val)
			}
			// Add process and function names as attributes
			cpuDataPoint.Attributes().PutStr("process.name", processName)
			cpuDataPoint.Attributes().PutStr("function.name", functionName)
			if filename, ok := functionToFilename[functionName]; ok && filename != "" {
				cpuDataPoint.Attributes().PutStr("file.name", filename)
				c.logDebug("Attached file.name to CPU datapoint",
					zap.String("process_name", processName),
					zap.String("function_name", functionName),
					zap.String("file_name", filename))
			}

			// Create memory data point with both process and function attributes
			memoryDataPoint := memoryGauge.DataPoints().AppendEmpty()
			memoryDataPoint.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
			memoryDataPoint.SetDoubleValue(memoryAllocation)

			// Add base attributes
			for key, val := range attributes {
				memoryDataPoint.Attributes().PutStr(key, val)
			}
			// Add process and function names as attributes
			memoryDataPoint.Attributes().PutStr("process.name", processName)
			memoryDataPoint.Attributes().PutStr("function.name", functionName)
			if filename, ok := functionToFilename[functionName]; ok && filename != "" {
				memoryDataPoint.Attributes().PutStr("file.name", filename)
				c.logDebug("Attached file.name to Memory datapoint",
					zap.String("process_name", processName),
					zap.String("function_name", functionName),
					zap.String("file_name", filename))
			}
		}
	}
}

// getUniqueFunctionNames extracts all unique function names from a profile
func (c *Converter) getUniqueFunctionNames(profiles pprofile.Profiles, profile pprofile.Profile) []string {
	c.logDebug("Starting to extract unique function names",
		zap.Int("samples_count", profile.Sample().Len()))

	functionNames := make(map[string]bool)

	for i := 0; i < profile.Sample().Len(); i++ {
		sample := profile.Sample().At(i)
		c.logDebug("Processing sample for function name",
			zap.Int("sample_index", i))

		functionName := c.getSampleFunctionName(profiles, sample)
		if functionName != "" {
			c.logDebug("Found function name",
				zap.Int("sample_index", i),
				zap.String("function_name", functionName))
			functionNames[functionName] = true
		} else {
			c.logDebug("Skipping sample with empty function name",
				zap.Int("sample_index", i))
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

// getFunctionFilenameMap builds a map from function name to source filename using the top location of samples
func (c *Converter) getFunctionFilenameMap(profiles pprofile.Profiles, profile pprofile.Profile) map[string]string {
	result := make(map[string]string)

	for i := 0; i < profile.Sample().Len(); i++ {
		sample := profile.Sample().At(i)
		functionName := c.getSampleFunctionName(profiles, sample)
		if functionName == "" {
			continue
		}

		// Resolve filename from the same top location
		filename := c.getSampleFileName(profiles, sample)
		c.logDebug("Resolved filename for function from sample",
			zap.Int("sample_index", i),
			zap.String("function_name", functionName),
			zap.String("file_name", filename))
		if filename == "" {
			continue
		}

		if _, exists := result[functionName]; !exists {
			result[functionName] = filename
		}
	}

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

		// Skip samples with empty function names
		if sampleFunctionName == "" {
			continue
		}

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

		// Skip samples with empty function names
		if sampleFunctionName == "" {
			continue
		}

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

// calculateFunctionCPUTimeForProcess calculates CPU time for a specific function within a specific process
func (c *Converter) calculateFunctionCPUTimeForProcess(
	profiles pprofile.Profiles,
	profile pprofile.Profile,
	processName, functionName string,
) float64 {
	var totalCPUTime float64
	defaultProfileDuration := 1.0
	sampleCount := profile.Sample().Len()

	for i := 0; i < sampleCount; i++ {
		sample := profile.Sample().At(i)

		// Check if sample belongs to this process
		sampleProcessName := c.getSampleAttributeValue(profiles, sample, "process.executable.name")
		if sampleProcessName != processName {
			continue
		}

		// Check if sample belongs to this function
		sampleFunctionName := c.getSampleFunctionName(profiles, sample)
		if sampleFunctionName == "" {
			continue // Skip samples with empty function names
		}
		if sampleFunctionName != functionName {
			continue
		}

		// Add the value
		values := sample.Values()
		if values.Len() > 0 {
			cpuTimeNs := float64(values.At(0))
			totalCPUTime += cpuTimeNs / nanosecondsPerSecond
		} else if sampleCount > 0 && defaultProfileDuration > 0 {
			totalCPUTime += defaultProfileDuration / float64(sampleCount)
		}
	}

	return totalCPUTime
}

// calculateFunctionMemoryAllocationForProcess calculates memory allocation for a specific function within a specific process
func (c *Converter) calculateFunctionMemoryAllocationForProcess(
	profiles pprofile.Profiles,
	profile pprofile.Profile,
	processName, functionName string,
) float64 {
	var totalMemoryAllocation float64
	sampleCount := profile.Sample().Len()

	for i := 0; i < sampleCount; i++ {
		sample := profile.Sample().At(i)

		// Check if sample belongs to this process
		sampleProcessName := c.getSampleAttributeValue(profiles, sample, "process.executable.name")
		if sampleProcessName != processName {
			continue
		}

		// Check if sample belongs to this function
		sampleFunctionName := c.getSampleFunctionName(profiles, sample)
		if sampleFunctionName == "" {
			continue // Skip samples with empty function names
		}
		if sampleFunctionName != functionName {
			continue
		}

		// Add the value
		values := sample.Values()
		if values.Len() > 1 {
			totalMemoryAllocation += float64(values.At(1))
		} else if values.Len() == 1 {
			totalMemoryAllocation += float64(values.At(0))
		} else {
			totalMemoryAllocation += 2048.0 // Default 2KB for stack trace profiles
		}
	}

	return totalMemoryAllocation
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
		c.logDebug("Function index is negative", zap.Int32("function_index", functionIndex))
		return ""
	}

	dictionary := profiles.Dictionary()
	functionTable := dictionary.FunctionTable()

	if int(functionIndex) >= functionTable.Len() {
		c.logDebug("Function index out of range",
			zap.Int32("function_index", functionIndex),
			zap.Int("function_table_len", functionTable.Len()))
		return ""
	}

	function := functionTable.At(int(functionIndex))
	nameIndex := function.NameStrindex()

	stringTable := dictionary.StringTable()
	if nameIndex < 0 || int(nameIndex) >= stringTable.Len() {
		c.logDebug("Function name index out of range",
			zap.Int32("name_index", nameIndex),
			zap.Int32("function_index", functionIndex),
			zap.Int("string_table_len", stringTable.Len()))
		return ""
	}

	functionName := stringTable.At(int(nameIndex))
	if functionName == "" {
		c.logDebug("Function name is empty string",
			zap.Int32("function_index", functionIndex),
			zap.Int32("name_index", nameIndex))
		return ""
	}

	c.logDebug("Resolved function name",
		zap.Int32("function_index", functionIndex),
		zap.String("function_name", functionName))

	return functionName
}

// getLocationFunctionName gets the function name from a location using the profiles dictionary
func (c *Converter) getLocationFunctionName(profiles pprofile.Profiles, location pprofile.Location) string {
	// Locations have Lines, and Lines have FunctionIndex
	lines := location.Line()
	c.logDebug("Getting function name from location",
		zap.Int("lines_count", lines.Len()))

	if lines.Len() == 0 {
		c.logDebug("Location has no lines")
		return ""
	}

	// Get the first line's function (most specific in the call stack)
	line := lines.At(0)
	functionIndex := line.FunctionIndex()

	c.logDebug("Location line info",
		zap.Int32("function_index", functionIndex))

	functionName := c.getFunctionName(profiles, functionIndex)
	if functionName == "" {
		c.logDebug("Function name is empty - skipping")
		return ""
	}

	return functionName
}

// getLocationFileName gets the source filename from a location using the profiles dictionary
func (c *Converter) getLocationFileName(profiles pprofile.Profiles, location pprofile.Location) string {
	filename := getLocationFileNameCommon(profiles, location)
	if filename == "" {
		c.logDebug("Location has no lines for filename resolution")
	} else {
		c.logDebug("Resolved file.name from location", zap.String("file_name", filename))
	}
	return filename
}

// getSampleFileName gets the top frame's source filename from a sample's stack
func (c *Converter) getSampleFileName(profiles pprofile.Profiles, sample pprofile.Sample) string {
	stackIndex := sample.StackIndex()
	if stackIndex < 0 {
		return ""
	}

	dictionary := profiles.Dictionary()
	stackTable := dictionary.StackTable()
	if int(stackIndex) >= stackTable.Len() {
		return ""
	}

	stack := stackTable.At(int(stackIndex))
	locationIndices := stack.LocationIndices()
	if locationIndices.Len() == 0 {
		return ""
	}

	locationIndex := locationIndices.At(locationIndices.Len() - 1)
	locationTable := dictionary.LocationTable()
	if locationIndex < 0 || int(locationIndex) >= locationTable.Len() {
		return ""
	}

	location := locationTable.At(int(locationIndex))
	filename := c.getLocationFileName(profiles, location)
	c.logDebug("Resolved file.name from sample top frame",
		zap.Int32("stack_index", stackIndex),
		zap.Int32("location_index", locationIndex),
		zap.String("file_name", filename))
	return filename
}

// getSampleFunctionName gets the top function name from a sample's stack
func (c *Converter) getSampleFunctionName(profiles pprofile.Profiles, sample pprofile.Sample) string {
	stackIndex := sample.StackIndex()
	c.logDebug("Getting function name from sample",
		zap.Int32("stack_index", stackIndex))

	if stackIndex < 0 {
		c.logDebug("Sample has no stack index")
		return ""
	}

	dictionary := profiles.Dictionary()
	stackTable := dictionary.StackTable()

	c.logDebug("Stack table info",
		zap.Int32("stack_index", stackIndex),
		zap.Int("stack_table_len", stackTable.Len()))

	if int(stackIndex) >= stackTable.Len() {
		c.logDebug("Stack index out of range",
			zap.Int32("stack_index", stackIndex),
			zap.Int("stack_table_len", stackTable.Len()))
		return ""
	}

	stack := stackTable.At(int(stackIndex))
	locationIndices := stack.LocationIndices()

	c.logDebug("Stack location indices",
		zap.Int32("stack_index", stackIndex),
		zap.Int("location_indices_count", locationIndices.Len()))

	if locationIndices.Len() == 0 {
		c.logDebug("Stack has no locations")
		return ""
	}

	// Get the LAST location (top of the call stack)
	// The stack grows downward, so the most recent function is at the end
	locationIndex := locationIndices.At(locationIndices.Len() - 1)
	locationTable := dictionary.LocationTable()

	c.logDebug("Location table info",
		zap.Int32("location_index", locationIndex),
		zap.Int("location_table_len", locationTable.Len()))

	if locationIndex < 0 || int(locationIndex) >= locationTable.Len() {
		c.logDebug("Location index out of range",
			zap.Int32("location_index", locationIndex),
			zap.Int("location_table_len", locationTable.Len()))
		return ""
	}

	location := locationTable.At(int(locationIndex))
	functionName := c.getLocationFunctionName(profiles, location)

	c.logDebug("Extracted function name from sample",
		zap.Int32("stack_index", stackIndex),
		zap.Int32("location_index", locationIndex),
		zap.String("function_name", functionName))

	return functionName
}

// getUniqueThreadNames extracts all unique thread names from a profile
// In the pprofile schema, thread information is stored as resource attributes
func (c *Converter) getUniqueThreadNames(profiles pprofile.Profiles, profile pprofile.Profile) []string {
	result := getUniqueAttributeValuesCommon(profiles, profile, "thread.name")
	c.logDebug("Extracted unique thread names", zap.Int("count", len(result)), zap.Strings("thread_names", result))
	return result
}

// getUniqueProcessNames extracts all unique process names from a profile
// In the pprofile schema, process information is stored as resource attributes
func (c *Converter) getUniqueProcessNames(profiles pprofile.Profiles, profile pprofile.Profile) []string {
	result := getUniqueAttributeValuesCommon(profiles, profile, "process.executable.name")
	c.logDebug("Extracted unique process names", zap.Int("count", len(result)), zap.Strings("process_names", result))
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
func (c *Converter) calculateMemoryAllocationForFilter(
	profiles pprofile.Profiles,
	profile pprofile.Profile,
	filter map[string]string,
) float64 {
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
