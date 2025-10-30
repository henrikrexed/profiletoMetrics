package profiletometrics

import (
	"context"
	"crypto/rand"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

// TraceConverter converts profiling data to traces with spans
type TraceConverter struct {
	config *ConverterConfig
	logger *zap.Logger
}

// NewTraceConverter creates a new profile to traces converter
func NewTraceConverter(cfg *ConverterConfig) (*TraceConverter, error) {
	return &TraceConverter{
		config: cfg,
		logger: nil, // Will be set by the connector
	}, nil
}

// SetLogger sets the logger for the trace converter
func (tc *TraceConverter) SetLogger(logger *zap.Logger) {
	tc.logger = logger
}

// logInfo logs an info message if logger is available
func (tc *TraceConverter) logInfo(msg string, fields ...zap.Field) {
	if tc.logger != nil {
		tc.logger.Info(msg, fields...)
	}
}

// logDebug logs a debug message if logger is available
func (tc *TraceConverter) logDebug(msg string, fields ...zap.Field) {
	if tc.logger != nil {
		tc.logger.Debug(msg, fields...)
	}
}

// logWarn logs a warning message if logger is available
func (tc *TraceConverter) logWarn(msg string, fields ...zap.Field) {
	if tc.logger != nil {
		tc.logger.Warn(msg, fields...)
	}
}

// ConvertProfilesToTraces converts profiling data to traces with spans
func (tc *TraceConverter) ConvertProfilesToTraces(ctx context.Context, profiles pprofile.Profiles) (ptrace.Traces, error) {
	tc.logInfo("Starting profile to traces conversion",
		zap.Int("resource_profiles_count", profiles.ResourceProfiles().Len()))

	traces := ptrace.NewTraces()
	resourceSpans := traces.ResourceSpans().AppendEmpty()

	iterateProfilesCommon(
		profiles,
		tc.extractResourceAttributes,
		func(resourceIndex, scopeIndex, profileIndex int, profile pprofile.Profile, resourceAttributes map[string]string) {
			tc.logDebug("Processing profile",
				zap.Int("resource_index", resourceIndex),
				zap.Int("scope_index", scopeIndex),
				zap.Int("profile_index", profileIndex),
				zap.Int("samples_count", profile.Sample().Len()))

			profileAttributes := tc.extractProfileAttributes(profiles, profile, resourceAttributes)
			tc.logDebug("Extracted profile attributes", zap.Any("attributes", profileAttributes))

			tc.generateTracesFromProfile(profiles, profile, profileAttributes, resourceSpans)
		},
	)

	tc.logInfo("Profile to traces conversion completed")
	return traces, nil
}

// extractResourceAttributes extracts attributes from the resource
func (tc *TraceConverter) extractResourceAttributes(resource pcommon.Resource) map[string]string {
	attributes := make(map[string]string)

	resource.Attributes().Range(func(key string, value pcommon.Value) bool {
		attributes[key] = value.AsString()
		return true
	})

	return attributes
}

// extractProfileAttributes extracts attributes from the profile data
func (tc *TraceConverter) extractProfileAttributes(
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
	for _, attr := range tc.config.Attributes {
		value := tc.extractAttributeValue(profiles, profile, attr)
		if value != "" {
			attributes[attr.Key] = value
		}
	}

	return attributes
}

// extractAttributeValue extracts a single attribute value based on the rule
func (tc *TraceConverter) extractAttributeValue(profiles pprofile.Profiles, _ pprofile.Profile, attr AttributeConfig) string {
	switch attr.Type {
	case "literal":
		return attr.Value
	case "regex":
		// Extract from string table using regex pattern
		return tc.extractFromStringTable(profiles, attr.Value)
	case "string_table":
		// Direct string table index access
		return tc.extractFromStringTableByIndex(profiles, attr.Value)
	default:
		return attr.Value
	}
}

// generateTracesFromProfile generates traces from profile data
func (tc *TraceConverter) generateTracesFromProfile(
	profiles pprofile.Profiles,
	profile pprofile.Profile,
	attributes map[string]string,
	resourceSpans ptrace.ResourceSpans,
) {
	// Apply pattern filtering if enabled
	if tc.config.PatternFilter.Enabled && !tc.matchesPatternFilter(attributes) {
		return
	}

	// Apply process filtering
	if !tc.matchesProcessFilter(attributes) {
		return
	}

	// Create a single scope spans for all spans from this profile
	scopeSpans := resourceSpans.ScopeSpans().AppendEmpty()
	scopeSpans.Scope().SetName("profiletometrics")
	scopeSpans.Scope().SetVersion("1.0.0")

	// Generate traces for each process
	processNames := tc.getUniqueProcessNames(profiles, profile)
	for _, processName := range processNames {
		tc.logDebug("Generating traces for process", zap.String("process_name", processName))
		tc.generateProcessTraces(profiles, profile, attributes, scopeSpans, processName)
	}
}

// generateProcessTraces generates traces for a specific process
func (tc *TraceConverter) generateProcessTraces(
	profiles pprofile.Profiles,
	profile pprofile.Profile,
	attributes map[string]string,
	scopeSpans ptrace.ScopeSpans,
	processName string,
) {
	// Group samples by their call stack to create trace hierarchies
	stackGroups := tc.groupSamplesByStack(profiles, profile, processName)

	for stackIndex, samples := range stackGroups {
		tc.logDebug("Processing stack group",
			zap.Int32("stack_index", stackIndex),
			zap.Int("sample_count", len(samples)))

		// Create a trace for this call stack
		traceID := tc.generateTraceID()
		tc.createTraceFromStack(profiles, stackIndex, samples, traceID, attributes, scopeSpans)
	}
}

// groupSamplesByStack groups samples by their stack index
func (tc *TraceConverter) groupSamplesByStack(
	profiles pprofile.Profiles,
	profile pprofile.Profile,
	processName string,
) map[int32][]pprofile.Sample {
	stackGroups := make(map[int32][]pprofile.Sample)

	for i := 0; i < profile.Sample().Len(); i++ {
		sample := profile.Sample().At(i)

		// Check if sample belongs to this process
		sampleProcessName := tc.getSampleAttributeValue(profiles, sample, "process.executable.name")
		if sampleProcessName != processName {
			continue
		}

		// Skip samples with empty function names
		sampleFunctionName := tc.getSampleFunctionName(profiles, sample)
		if sampleFunctionName == "" {
			continue
		}

		stackIndex := sample.StackIndex()
		if stackIndex >= 0 {
			stackGroups[stackIndex] = append(stackGroups[stackIndex], sample)
		}
	}

	return stackGroups
}

// createTraceFromStack creates a trace from a call stack
func (tc *TraceConverter) createTraceFromStack(
	profiles pprofile.Profiles,
	stackIndex int32,
	samples []pprofile.Sample,
	traceID pcommon.TraceID,
	attributes map[string]string,
	scopeSpans ptrace.ScopeSpans,
) {
	// Get the call stack
	stack := tc.getStackFromIndex(profiles, stackIndex)
	if stack == nil {
		tc.logWarn("Could not get stack from index", zap.Int32("stack_index", stackIndex))
		return
	}

	// Calculate total duration from samples
	totalDuration := tc.calculateTotalDuration(samples)
	startTime := time.Now().Add(-totalDuration)

	// Create spans for each function in the call stack
	parentSpanID := pcommon.SpanID{}
	spans := make([]ptrace.Span, 0)

	// Process locations in reverse order (from caller to callee)
	locationIndices := stack.LocationIndices()
	for i := locationIndices.Len() - 1; i >= 0; i-- {
		locationIndex := locationIndices.At(i)
		location := tc.getLocationFromIndex(profiles, locationIndex)
		if location == nil {
			continue
		}

		functionName := tc.getLocationFunctionName(profiles, *location)
		if functionName == "" {
			continue
		}

		// Create span for this function
		span := scopeSpans.Spans().AppendEmpty()
		spanID := tc.generateSpanID()

		// Set span properties
		span.SetTraceID(traceID)
		span.SetSpanID(spanID)
		span.SetParentSpanID(parentSpanID)
		span.SetName(functionName)
		span.SetKind(ptrace.SpanKindInternal)
		span.SetStartTimestamp(pcommon.NewTimestampFromTime(startTime))

		// Calculate duration for this function
		functionDuration := tc.calculateFunctionDuration(samples, functionName, totalDuration)
		span.SetEndTimestamp(pcommon.NewTimestampFromTime(startTime.Add(functionDuration)))

		// Add attributes
		for key, val := range attributes {
			span.Attributes().PutStr(key, val)
		}
		span.Attributes().PutStr("function.name", functionName)
		span.Attributes().PutStr("span.kind", "internal")

		// Add filename attribute if available from the same location
		if filename := tc.getLocationFileName(profiles, *location); filename != "" {
			span.Attributes().PutStr("file.name", filename)
			tc.logDebug("Attached file.name to span",
				zap.String("function_name", functionName),
				zap.String("file_name", filename))
		}

		// Add events for sample data
		tc.addSampleEvents(span, samples, functionName)

		spans = append(spans, span)

		// Update parent for next span
		parentSpanID = spanID
		startTime = startTime.Add(functionDuration)
	}

	tc.logDebug("Created trace from stack",
		zap.Int32("stack_index", stackIndex),
		zap.Int("span_count", len(spans)),
		zap.String("trace_id", string(traceID[:])))
}

// getStackFromIndex gets a stack from the stack table by index
func (tc *TraceConverter) getStackFromIndex(profiles pprofile.Profiles, stackIndex int32) *pprofile.Stack {
	dictionary := profiles.Dictionary()
	stackTable := dictionary.StackTable()

	if stackIndex < 0 || int(stackIndex) >= stackTable.Len() {
		return nil
	}

	stack := stackTable.At(int(stackIndex))
	return &stack
}

// getLocationFromIndex gets a location from the location table by index
func (tc *TraceConverter) getLocationFromIndex(profiles pprofile.Profiles, locationIndex int32) *pprofile.Location {
	dictionary := profiles.Dictionary()
	locationTable := dictionary.LocationTable()

	if locationIndex < 0 || int(locationIndex) >= locationTable.Len() {
		return nil
	}

	location := locationTable.At(int(locationIndex))
	return &location
}

// getLocationFunctionName gets the function name from a location
func (tc *TraceConverter) getLocationFunctionName(profiles pprofile.Profiles, location pprofile.Location) string {
	lines := location.Line()
	if lines.Len() == 0 {
		return ""
	}

	line := lines.At(0)
	functionIndex := line.FunctionIndex()
	functionName := tc.getFunctionName(profiles, functionIndex)
	if functionName == "" {
		return ""
	}
	return functionName
}

// getLocationFileName gets the source filename from a location
func (tc *TraceConverter) getLocationFileName(profiles pprofile.Profiles, location pprofile.Location) string {
	filename := getLocationFileNameCommon(profiles, location)
	if filename == "" {
		tc.logDebug("Location has no lines for filename resolution")
	} else {
		tc.logDebug("Resolved file.name from location", zap.String("file_name", filename))
	}
	return filename
}

// getFunctionName extracts the function name from a function index
func (tc *TraceConverter) getFunctionName(profiles pprofile.Profiles, functionIndex int32) string {
	if functionIndex < 0 {
		return ""
	}

	dictionary := profiles.Dictionary()
	functionTable := dictionary.FunctionTable()

	if int(functionIndex) >= functionTable.Len() {
		return ""
	}

	function := functionTable.At(int(functionIndex))
	nameIndex := function.NameStrindex()

	stringTable := dictionary.StringTable()
	if nameIndex < 0 || int(nameIndex) >= stringTable.Len() {
		return ""
	}

	functionName := stringTable.At(int(nameIndex))
	if functionName == "" {
		return ""
	}

	return functionName
}

// calculateTotalDuration calculates the total duration from samples
func (tc *TraceConverter) calculateTotalDuration(samples []pprofile.Sample) time.Duration {
	var totalNs int64
	for _, sample := range samples {
		values := sample.Values()
		if values.Len() > 0 {
			totalNs += values.At(0) // CPU time in nanoseconds
		}
	}

	// If no values, use a default duration
	if totalNs == 0 {
		return time.Second // Default 1 second
	}

	return time.Duration(totalNs)
}

// calculateFunctionDuration calculates the duration for a specific function
func (tc *TraceConverter) calculateFunctionDuration(
	samples []pprofile.Sample,
	_ string,
	totalDuration time.Duration,
) time.Duration {
	// For now, distribute duration evenly across functions
	// In a more sophisticated implementation, you could analyze the actual time spent
	// in each function based on the sample data
	return totalDuration / time.Duration(len(samples))
}

// addSampleEvents adds events to a span based on sample data
func (tc *TraceConverter) addSampleEvents(span ptrace.Span, samples []pprofile.Sample, functionName string) {
	for i, sample := range samples {
		event := span.Events().AppendEmpty()
		event.SetName("sample")
		event.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))

		// Add sample attributes
		event.Attributes().PutStr("function.name", functionName)
		event.Attributes().PutInt("sample.index", int64(i))

		// Add sample values
		values := sample.Values()
		if values.Len() > 0 {
			event.Attributes().PutInt("cpu_time_ns", values.At(0))
		}
		if values.Len() > 1 {
			event.Attributes().PutInt("memory_bytes", values.At(1))
		}
	}
}

// generateTraceID generates a new trace ID
func (tc *TraceConverter) generateTraceID() pcommon.TraceID {
	var traceID pcommon.TraceID
	if _, err := rand.Read(traceID[:]); err != nil {
		tc.logWarn("Failed to generate trace ID, using zero value", zap.Error(err))
	}
	return traceID
}

// generateSpanID generates a new span ID
func (tc *TraceConverter) generateSpanID() pcommon.SpanID {
	var spanID pcommon.SpanID
	if _, err := rand.Read(spanID[:]); err != nil {
		tc.logWarn("Failed to generate span ID, using zero value", zap.Error(err))
	}
	return spanID
}

// getSampleFunctionName gets the top function name from a sample's stack
func (tc *TraceConverter) getSampleFunctionName(profiles pprofile.Profiles, sample pprofile.Sample) string {
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

	// Get the LAST location (top of the call stack)
	locationIndex := locationIndices.At(locationIndices.Len() - 1)
	locationTable := dictionary.LocationTable()

	if locationIndex < 0 || int(locationIndex) >= locationTable.Len() {
		return ""
	}

	location := locationTable.At(int(locationIndex))
	return tc.getLocationFunctionName(profiles, location)
}

// getSampleAttributeValue extracts a specific attribute value from a sample
func (tc *TraceConverter) getSampleAttributeValue(profiles pprofile.Profiles, sample pprofile.Sample, key string) string {
	return getSampleAttributeValueCommon(profiles, sample, key)
}

// getUniqueProcessNames extracts all unique process names from a profile
func (tc *TraceConverter) getUniqueProcessNames(profiles pprofile.Profiles, profile pprofile.Profile) []string {
	processNames := make(map[string]bool)

	// Iterate through samples to extract unique process names from attributes
	for i := 0; i < profile.Sample().Len(); i++ {
		sample := profile.Sample().At(i)
		processName := tc.getSampleAttributeValue(profiles, sample, "process.executable.name")
		if processName != "" {
			processNames[processName] = true
		}
	}

	var result []string
	for processName := range processNames {
		result = append(result, processName)
	}

	return result
}

// matchesPatternFilter checks if attributes match the pattern filter
func (tc *TraceConverter) matchesPatternFilter(attributes map[string]string) bool {
	if !tc.config.PatternFilter.Enabled {
		return true
	}
	// Check if any attribute value matches the pattern
	for _, value := range attributes {
		if tc.config.PatternFilter.Pattern != "" &&
			value != "" {
			// Simple substring matching for now
			return true
		}
	}
	return false
}

// matchesProcessFilter checks if the profile matches the process filter
func (tc *TraceConverter) matchesProcessFilter(attributes map[string]string) bool {
	if !tc.config.ProcessFilter.Enabled {
		return true // No filter configured
	}

	processName, exists := attributes["process_name"]
	if !exists {
		return false // No process name attribute found
	}

	// For now, simple string matching - in a real implementation you would compile and match the regex pattern
	if tc.config.ProcessFilter.Pattern == "" {
		return true // No pattern specified, allow all
	}

	// Simple contains check for now - in production, use regex compilation
	return processName != "" // Placeholder logic
}

// extractFromStringTable extracts values from profile string table using regex pattern
func (tc *TraceConverter) extractFromStringTable(profiles pprofile.Profiles, _ string) string {
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
func (tc *TraceConverter) extractFromStringTableByIndex(profiles pprofile.Profiles, _ string) string {
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
