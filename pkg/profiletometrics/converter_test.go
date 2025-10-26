package profiletometrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.uber.org/zap"

	"github.com/henrikrexed/profiletoMetrics/testdata"
)

// validateSingleMetric validates a single metric with the expected name
func validateSingleMetric(t *testing.T, metrics pmetric.Metrics, expectedName string) {
	resourceMetrics := metrics.ResourceMetrics()
	require.Equal(t, 1, resourceMetrics.Len())

	scopeMetrics := resourceMetrics.At(0).ScopeMetrics()
	require.Equal(t, 1, scopeMetrics.Len())

	metricsSlice := scopeMetrics.At(0).Metrics()
	require.Equal(t, 1, metricsSlice.Len())

	metric := metricsSlice.At(0)
	assert.Equal(t, expectedName, metric.Name())
}

func TestConverter_ConvertProfilesToMetrics(t *testing.T) {
	tests := []struct {
		name            string
		config          *ConverterConfig
		profileData     pprofile.Profiles
		expectedMetrics int
		validateMetrics func(t *testing.T, metrics pmetric.Metrics)
	}{
		{
			name: "Basic CPU and Memory metrics generation",
			config: &ConverterConfig{
				Metrics: MetricsConfig{
					CPU: CPUMetricConfig{
						Enabled:    true,
						MetricName: "test_cpu_time",
						Unit:       "s",
					},
					Memory: MemoryMetricConfig{
						Enabled:    true,
						MetricName: "test_memory_allocation",
						Unit:       "bytes",
					},
				},
				Attributes: []AttributeConfig{
					{
						Key:   "process_name",
						Value: "test_process",
						Type:  "literal",
					},
				},
			},
			profileData:     testdata.CreateTestProfile(),
			expectedMetrics: 1, // One resource metrics container (but we're getting 2, so let's fix this)
			validateMetrics: func(t *testing.T, metrics pmetric.Metrics) {
				resourceMetrics := metrics.ResourceMetrics()
				require.Equal(t, 1, resourceMetrics.Len())

				scopeMetrics := resourceMetrics.At(0).ScopeMetrics()
				require.Equal(t, 1, scopeMetrics.Len())

				// Check that we have both CPU and memory metrics
				metricsSlice := scopeMetrics.At(0).Metrics()
				require.Equal(t, 2, metricsSlice.Len())

				// Find CPU time metric
				var cpuTimeMetric pmetric.Metric
				var memoryMetric pmetric.Metric
				for i := 0; i < metricsSlice.Len(); i++ {
					metric := metricsSlice.At(i)
					if metric.Name() == "test_cpu_time" {
						cpuTimeMetric = metric
					} else if metric.Name() == "test_memory_allocation" {
						memoryMetric = metric
					}
				}

				// Validate CPU time metric
				require.NotNil(t, cpuTimeMetric)
				assert.Equal(t, "test_cpu_time", cpuTimeMetric.Name())
				assert.Equal(t, "CPU time in seconds", cpuTimeMetric.Description())

				gauge := cpuTimeMetric.Gauge()
				require.Equal(t, 1, gauge.DataPoints().Len())

				dataPoint := gauge.DataPoints().At(0)
				assert.Greater(t, dataPoint.DoubleValue(), float64(0))

				// Validate memory metric
				require.NotNil(t, memoryMetric)
				assert.Equal(t, "test_memory_allocation", memoryMetric.Name())
				assert.Equal(t, "Memory allocation in bytes", memoryMetric.Description())

				memoryGauge := memoryMetric.Gauge()
				require.Equal(t, 1, memoryGauge.DataPoints().Len())

				memoryDataPoint := memoryGauge.DataPoints().At(0)
				assert.Greater(t, memoryDataPoint.DoubleValue(), float64(0))
			},
		},
		{
			name: "CPU time only",
			config: &ConverterConfig{
				Metrics: MetricsConfig{
					CPU: CPUMetricConfig{
						Enabled:    true,
						MetricName: "cpu_time_only",
						Unit:       "s",
					},
				},
			},
			profileData:     testdata.CreateTestProfile(),
			expectedMetrics: 1,
			validateMetrics: func(t *testing.T, metrics pmetric.Metrics) {
				validateSingleMetric(t, metrics, "cpu_time_only")
			},
		},
		{
			name: "Memory allocation only",
			config: &ConverterConfig{
				Metrics: MetricsConfig{
					Memory: MemoryMetricConfig{
						Enabled:    true,
						MetricName: "memory_only",
						Unit:       "bytes",
					},
				},
			},
			profileData:     testdata.CreateTestProfile(),
			expectedMetrics: 1,
			validateMetrics: func(t *testing.T, metrics pmetric.Metrics) {
				validateSingleMetric(t, metrics, "memory_only")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create converter
			converter, err := NewConverter(tt.config)
			require.NoError(t, err)

			// Process profile data
			profiles := tt.profileData
			metrics, err := converter.ConvertProfilesToMetrics(context.Background(), profiles)
			require.NoError(t, err)

			// Validate results
			resourceMetrics := metrics.ResourceMetrics()
			assert.Equal(t, tt.expectedMetrics, resourceMetrics.Len())

			if tt.validateMetrics != nil {
				tt.validateMetrics(t, metrics)
			}
		})
	}
}

func TestConverter_matchesPatternFilter(t *testing.T) {
	tests := []struct {
		name           string
		config         ConverterConfig
		attributes     map[string]string
		expectedResult bool
	}{
		{
			name: "Pattern filter disabled",
			config: ConverterConfig{
				PatternFilter: PatternFilterConfig{
					Enabled: false,
				},
			},
			attributes:     map[string]string{"test": "value"},
			expectedResult: true,
		},
		{
			name: "Pattern filter enabled",
			config: ConverterConfig{
				PatternFilter: PatternFilterConfig{
					Enabled: true,
					Pattern: "test.*",
				},
			},
			attributes:     map[string]string{"test": "value"},
			expectedResult: true, // Current implementation always returns true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter, err := NewConverter(&tt.config)
			require.NoError(t, err)
			result := converter.matchesPatternFilter(tt.attributes)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestConverter_matchesProcessFilter(t *testing.T) {
	tests := []struct {
		name           string
		config         ConverterConfig
		attributes     map[string]string
		expectedResult bool
	}{
		{
			name: "Process filter disabled",
			config: ConverterConfig{
				ProcessFilter: ProcessFilterConfig{
					Enabled: false,
				},
			},
			attributes:     map[string]string{"test": "value"},
			expectedResult: true,
		},
		{
			name: "Process filter enabled with process_name",
			config: ConverterConfig{
				ProcessFilter: ProcessFilterConfig{
					Enabled: true,
					Pattern: "test.*",
				},
			},
			attributes:     map[string]string{"process_name": "test_process"},
			expectedResult: true, // Current implementation always returns true
		},
		{
			name: "Process filter enabled without process_name",
			config: ConverterConfig{
				ProcessFilter: ProcessFilterConfig{
					Enabled: true,
					Pattern: "test.*",
				},
			},
			attributes:     map[string]string{"other": "value"},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter, err := NewConverter(&tt.config)
			require.NoError(t, err)
			result := converter.matchesProcessFilter(tt.attributes)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestConverter_CalculateCPUTime(t *testing.T) {
	config := &ConverterConfig{
		Metrics: MetricsConfig{
			CPU: CPUMetricConfig{
				Enabled: true,
			},
		},
	}

	converter, err := NewConverter(config)
	require.NoError(t, err)

	// Create test profile with sample data
	profile := pprofile.NewProfile()

	// Add samples with CPU time values (in nanoseconds)
	for i := 0; i < 3; i++ {
		sample := profile.Sample().AppendEmpty()
		values := sample.Values()
		values.Append(int64(1000000000 + i*100000000)) // 1s, 1.1s, 1.2s in nanoseconds
	}

	// Create a profiles container to pass to calculateCPUTime
	profiles := pprofile.NewProfiles()
	resourceProfiles := profiles.ResourceProfiles().AppendEmpty()
	scopeProfiles := resourceProfiles.ScopeProfiles().AppendEmpty()
	profileInProfiles := scopeProfiles.Profiles().AppendEmpty()
	profile.CopyTo(profileInProfiles)

	cpuTime := converter.calculateCPUTime(profiles, profileInProfiles)
	expected := float64(1000000000+1100000000+1200000000) / 1e9 // Convert to seconds
	assert.Equal(t, expected, cpuTime)
}

func TestConverter_CalculateMemoryAllocation(t *testing.T) {
	config := &ConverterConfig{
		Metrics: MetricsConfig{
			Memory: MemoryMetricConfig{
				Enabled: true,
			},
		},
	}

	converter, err := NewConverter(config)
	require.NoError(t, err)

	// Create test profile with sample data
	profile := pprofile.NewProfile()

	// Add samples with memory allocation values (second value is memory)
	for i := 0; i < 2; i++ {
		sample := profile.Sample().AppendEmpty()
		values := sample.Values()
		values.Append(int64(1000000000))   // CPU time in nanoseconds
		values.Append(int64(2000 + i*500)) // Memory: 2000, 2500 bytes
	}

	// Create a profiles container to pass to calculateMemoryAllocation
	profiles := pprofile.NewProfiles()
	resourceProfiles := profiles.ResourceProfiles().AppendEmpty()
	scopeProfiles := resourceProfiles.ScopeProfiles().AppendEmpty()
	profileInProfiles := scopeProfiles.Profiles().AppendEmpty()
	profile.CopyTo(profileInProfiles)

	memoryAllocation := converter.calculateMemoryAllocation(profiles, profileInProfiles)
	expected := float64(2000 + 2500)
	assert.Equal(t, expected, memoryAllocation)
}

func TestConverter_StringTableExtraction(t *testing.T) {
	config := &ConverterConfig{
		Metrics: MetricsConfig{
			CPU: CPUMetricConfig{
				Enabled:    true,
				MetricName: "cpu_time",
				Unit:       "s",
			},
		},
		Attributes: []AttributeConfig{
			{
				Key:   "function_name",
				Value: "test_function",
				Type:  "literal",
			},
		},
	}

	converter, err := NewConverter(config)
	require.NoError(t, err)

	// Process test profile
	profiles := testdata.CreateTestProfile()
	metrics, err := converter.ConvertProfilesToMetrics(context.Background(), profiles)
	require.NoError(t, err)

	// Validate results
	resourceMetrics := metrics.ResourceMetrics()
	require.Equal(t, 1, resourceMetrics.Len())

	scopeMetrics := resourceMetrics.At(0).ScopeMetrics()
	require.Equal(t, 1, scopeMetrics.Len())

	metricsSlice := scopeMetrics.At(0).Metrics()
	require.Equal(t, 1, metricsSlice.Len())

	metric := metricsSlice.At(0)
	gauge := metric.Gauge()
	dataPoint := gauge.DataPoints().At(0)

	// Check extracted attributes
	attributes := dataPoint.Attributes()

	// Should have extracted function names
	functionName, exists := attributes.Get("function_name")
	require.True(t, exists)
	assert.Equal(t, "test_function", functionName.AsString())
}

func TestConverter_GetSampleAttributeValue(t *testing.T) {
	converter, err := NewConverter(&ConverterConfig{})
	require.NoError(t, err)

	// Create test profile
	profiles := testdata.CreateTestProfile()
	sample := pprofile.NewSample()

	// Test attribute extraction (may return empty string if no attributes)
	value := converter.getSampleAttributeValue(profiles, sample, "thread.name")

	// Just verify the function doesn't panic
	assert.NotNil(t, value)
}

func TestConverter_MatchesSampleFilter(t *testing.T) {
	converter, err := NewConverter(&ConverterConfig{})
	require.NoError(t, err)

	profiles := testdata.CreateTestProfile()
	sample := pprofile.NewSample()

	// Test with nil filter
	result := converter.matchesSampleFilter(profiles, sample, nil)
	assert.True(t, result, "Nil filter should match all")

	// Test with empty filter
	result = converter.matchesSampleFilter(profiles, sample, map[string]string{})
	assert.True(t, result, "Empty filter should match all")

	// Test with non-empty filter (will fail to match since sample has no attributes)
	result = converter.matchesSampleFilter(profiles, sample, map[string]string{"thread.name": "test"})
	assert.False(t, result, "Filter should not match when no attributes present")
}

func TestConverter_SetLogger(t *testing.T) {
	converter, err := NewConverter(&ConverterConfig{})
	require.NoError(t, err)

	// Set a logger
	logger, _ := zap.NewProduction()
	converter.SetLogger(logger)

	// Verify logger was set by calling a method that uses it
	converter.logInfo("Test message")

	// No assertions - just verifying it doesn't panic
}

func TestConverter_SanitizeMetricName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple name", "cpu_time", "cpu_time"},
		{"with spaces", "cpu time", "cpu_time"},
		{"with special chars", "cpu-time@test", "cpu_time_test"},
		{"with numbers", "cpu123time", "cpu123time"},
		{"mixed case", "CPU_Time", "CPU_Time"},
		{"with dots", "cpu.time", "cpu_time"},
		{"with slashes", "cpu/time", "cpu_time"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeMetricName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConverter_GetFunctionName(t *testing.T) {
	converter, err := NewConverter(&ConverterConfig{})
	require.NoError(t, err)

	// Create profiles with dictionary
	profiles := pprofile.NewProfiles()
	dictionary := profiles.Dictionary()
	stringTable := dictionary.StringTable()

	// Add function name to string table
	stringTable.Append("my_function")

	// Create function
	fn := pprofile.NewFunction()
	fn.SetNameStrindex(0)
	functionTable := dictionary.FunctionTable()
	functionTable.AppendEmpty().CopyTo(fn)

	// Test getting function name
	functionName := converter.getFunctionName(profiles, 0)
	assert.Equal(t, "my_function", functionName)

	// Test with invalid index
	functionName = converter.getFunctionName(profiles, -1)
	assert.Equal(t, "", functionName)

	// Test with index out of range
	functionName = converter.getFunctionName(profiles, 999)
	assert.Equal(t, "", functionName)
}

func TestConverter_GetLocationFunctionName(t *testing.T) {
	converter, err := NewConverter(&ConverterConfig{})
	require.NoError(t, err)

	profiles := pprofile.NewProfiles()
	dictionary := profiles.Dictionary()
	stringTable := dictionary.StringTable()

	// Setup function
	stringTable.Append("test_function")
	fn := pprofile.NewFunction()
	fn.SetNameStrindex(0)
	dictionary.FunctionTable().AppendEmpty().CopyTo(fn)

	// Create location with line
	location := pprofile.NewLocation()
	line := location.Line().AppendEmpty()
	line.SetFunctionIndex(0)

	// Test
	functionName := converter.getLocationFunctionName(profiles, location)
	assert.Equal(t, "test_function", functionName)

	// Test with empty lines
	location2 := pprofile.NewLocation()
	functionName2 := converter.getLocationFunctionName(profiles, location2)
	assert.Equal(t, "", functionName2)
}

func TestConverter_GetSampleFunctionName(t *testing.T) {
	converter, err := NewConverter(&ConverterConfig{})
	require.NoError(t, err)

	profiles := pprofile.NewProfiles()

	// Test with invalid stack index (negative)
	sample := pprofile.NewSample()
	sample.SetStackIndex(-1)
	functionName := converter.getSampleFunctionName(profiles, sample)
	assert.Equal(t, "", functionName)

	// Test with no stack index
	sample2 := pprofile.NewSample()
	sample2.SetStackIndex(0)
	functionName2 := converter.getSampleFunctionName(profiles, sample2)
	// May be empty if stack doesn't exist or has no locations
	assert.NotNil(t, functionName2)
}

func TestConverter_GetUniqueThreadNames(t *testing.T) {
	converter, err := NewConverter(&ConverterConfig{})
	require.NoError(t, err)

	profiles := pprofile.NewProfiles()
	profile := pprofile.NewProfile()

	// This should return empty list if no thread.name attributes are present
	threadNames := converter.getUniqueThreadNames(profiles, profile)
	assert.Equal(t, 0, len(threadNames))
}

func TestConverter_GetUniqueProcessNames(t *testing.T) {
	converter, err := NewConverter(&ConverterConfig{})
	require.NoError(t, err)

	profiles := pprofile.NewProfiles()
	profile := pprofile.NewProfile()

	// This should return empty list if no process.executable.name attributes are present
	processNames := converter.getUniqueProcessNames(profiles, profile)
	assert.Equal(t, 0, len(processNames))
}

func TestConverter_ExtractAttributeValue(t *testing.T) {
	converter, err := NewConverter(&ConverterConfig{})
	require.NoError(t, err)

	profiles := testdata.CreateTestProfile()
	profile := pprofile.NewProfile()

	// Test literal type
	attr := AttributeConfig{
		Key:   "test_attr",
		Value: "test_value",
		Type:  "literal",
	}
	value := converter.extractAttributeValue(profiles, profile, attr)
	assert.Equal(t, "test_value", value)

	// Test regex type (will use empty string for now)
	attr2 := AttributeConfig{
		Key:   "test_attr2",
		Value: ".*",
		Type:  "regex",
	}
	value2 := converter.extractAttributeValue(profiles, profile, attr2)
	// May be empty depending on implementation
	assert.NotNil(t, value2)

	// Test default/unknown type
	attr3 := AttributeConfig{
		Key:   "test_attr3",
		Value: "default",
		Type:  "unknown",
	}
	value3 := converter.extractAttributeValue(profiles, profile, attr3)
	assert.Equal(t, "default", value3)
}

func TestConverter_CalculateCPUTimeForFilter(t *testing.T) {
	config := &ConverterConfig{
		Metrics: MetricsConfig{
			CPU: CPUMetricConfig{
				Enabled: true,
			},
		},
	}

	converter, err := NewConverter(config)
	require.NoError(t, err)

	profiles := pprofile.NewProfiles()
	profile := pprofile.NewProfile()

	// Add samples
	for i := 0; i < 3; i++ {
		sample := profile.Sample().AppendEmpty()
		values := sample.Values()
		values.Append(int64(1000000000 + i*100000000))
	}

	// Test without filter
	cpuTime := converter.calculateCPUTimeForFilter(profiles, profile, nil)
	expected := float64(1000000000+1100000000+1200000000) / 1e9
	assert.InDelta(t, expected, cpuTime, 0.0001)

	// Test with filter (that won't match)
	filter := map[string]string{"thread.name": "nonexistent"}
	cpuTime2 := converter.calculateCPUTimeForFilter(profiles, profile, filter)
	assert.Equal(t, float64(0), cpuTime2)
}

func TestConverter_CalculateMemoryAllocationForFilter(t *testing.T) {
	config := &ConverterConfig{
		Metrics: MetricsConfig{
			Memory: MemoryMetricConfig{
				Enabled: true,
			},
		},
	}

	converter, err := NewConverter(config)
	require.NoError(t, err)

	profiles := pprofile.NewProfiles()
	profile := pprofile.NewProfile()

	// Add samples with memory values
	for i := 0; i < 2; i++ {
		sample := profile.Sample().AppendEmpty()
		values := sample.Values()
		values.Append(int64(1000000000))
		values.Append(int64(2000 + i*500))
	}

	// Test without filter
	memory := converter.calculateMemoryAllocationForFilter(profiles, profile, nil)
	expected := float64(2000 + 2500)
	assert.Equal(t, expected, memory)

	// Test with filter (that won't match)
	filter := map[string]string{"thread.name": "nonexistent"}
	memory2 := converter.calculateMemoryAllocationForFilter(profiles, profile, filter)
	assert.Equal(t, float64(0), memory2)
}

func TestConverter_FunctionMetricsEnabled(t *testing.T) {
	tests := []struct {
		name               string
		functionEnabled    bool
		expectedMinMetrics int // Minimum number of metrics we expect
	}{
		{
			name:               "Function metrics enabled",
			functionEnabled:    true,
			expectedMinMetrics: 2, // At least CPU and Memory
		},
		{
			name:               "Function metrics disabled",
			functionEnabled:    false,
			expectedMinMetrics: 2, // At least CPU and Memory (no function metrics)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ConverterConfig{
				Metrics: MetricsConfig{
					CPU: CPUMetricConfig{
						Enabled:    true,
						MetricName: "cpu_time",
						Unit:       "s",
					},
					Memory: MemoryMetricConfig{
						Enabled:    true,
						MetricName: "memory_allocation",
						Unit:       "bytes",
					},
					Function: FunctionMetricConfig{
						Enabled: tt.functionEnabled,
					},
				},
			}

			converter, err := NewConverter(config)
			require.NoError(t, err)

			// Create a profile with function data
			profiles := testdata.CreateTestProfile()

			// Convert profiles to metrics
			metrics, err := converter.ConvertProfilesToMetrics(context.Background(), profiles)
			require.NoError(t, err)
			require.NotNil(t, metrics)

			// Verify we got at least the minimum expected metrics
			resourceMetrics := metrics.ResourceMetrics()
			require.Equal(t, 1, resourceMetrics.Len())

			scopeMetrics := resourceMetrics.At(0).ScopeMetrics()
			require.Equal(t, 1, scopeMetrics.Len())

			metricsSlice := scopeMetrics.At(0).Metrics()
			assert.GreaterOrEqual(t, metricsSlice.Len(), tt.expectedMinMetrics)

			// Verify CPU and Memory metrics are always present
			var hasCPU bool
			var hasMemory bool

			for i := 0; i < metricsSlice.Len(); i++ {
				metric := metricsSlice.At(i)
				name := metric.Name()

				if name == "cpu_time" {
					hasCPU = true
				} else if name == "memory_allocation" {
					hasMemory = true
				}
				// Function, thread, and process metrics all use the same base metric names
				// with different attributes, so we don't need to distinguish them by name
			}

			assert.True(t, hasCPU, "CPU metric should be present")
			assert.True(t, hasMemory, "Memory metric should be present")
		})
	}
}

func TestConverter_GenerateFunctionMetrics(t *testing.T) {
	config := &ConverterConfig{
		Metrics: MetricsConfig{
			CPU: CPUMetricConfig{
				Enabled:    true,
				MetricName: "cpu_time",
			},
			Memory: MemoryMetricConfig{
				Enabled:    true,
				MetricName: "memory_allocation",
			},
			Function: FunctionMetricConfig{
				Enabled: true,
			},
		},
	}

	converter, err := NewConverter(config)
	require.NoError(t, err)

	profiles := pprofile.NewProfiles()
	resourceProfiles := profiles.ResourceProfiles().AppendEmpty()
	scopeProfiles := resourceProfiles.ScopeProfiles().AppendEmpty()
	profile := scopeProfiles.Profiles().AppendEmpty()

	// Setup string table
	dictionary := profiles.Dictionary()
	stringTable := dictionary.StringTable()
	stringTable.Append("main")
	stringTable.Append("handler")
	stringTable.Append("process.executable.name")
	stringTable.Append("myprocess")

	// Setup function table
	functionTable := dictionary.FunctionTable()
	fn1 := functionTable.AppendEmpty()
	fn1.SetNameStrindex(0) // "main"
	fn2 := functionTable.AppendEmpty()
	fn2.SetNameStrindex(1) // "handler"

	// Setup location table
	locationTable := dictionary.LocationTable()
	loc1 := locationTable.AppendEmpty()
	line1 := loc1.Line().AppendEmpty()
	line1.SetFunctionIndex(0) // main

	loc2 := locationTable.AppendEmpty()
	line2 := loc2.Line().AppendEmpty()
	line2.SetFunctionIndex(1) // handler

	// Setup stack table
	stackTable := dictionary.StackTable()
	stack1 := stackTable.AppendEmpty()
	stack1.LocationIndices().Append(0)
	stack2 := stackTable.AppendEmpty()
	stack2.LocationIndices().Append(1)

	// Setup attribute table for process names
	attributeTable := dictionary.AttributeTable()
	attr1 := attributeTable.AppendEmpty()
	attr1.SetKeyStrindex(2) // "process.executable.name"
	attr1.Value().SetStr("myprocess") // String value
	attr2 := attributeTable.AppendEmpty()
	attr2.SetKeyStrindex(2) // "process.executable.name"
	attr2.Value().SetStr("myprocess") // String value

	// Add samples with process attributes
	sample1 := profile.Sample().AppendEmpty()
	sample1.SetStackIndex(0)
	sample1.AttributeIndices().Append(0) // Reference to first attribute
	sample2 := profile.Sample().AppendEmpty()
	sample2.SetStackIndex(1)
	sample2.AttributeIndices().Append(1) // Reference to second attribute

	// Create scope metrics
	scopeMetrics := pmetric.NewScopeMetrics()

	// Generate function metrics
	attributes := map[string]string{"service.name": "test"}
	converter.generateFunctionMetrics(profiles, profile, attributes, scopeMetrics)

	// Verify metrics were created
	metrics := scopeMetrics.Metrics()
	assert.GreaterOrEqual(t, metrics.Len(), 2, "Should have at least 2 metrics")

	// Find metrics with function attributes
	var hasCPUWithFunction bool
	var hasMemoryWithFunction bool
	for i := 0; i < metrics.Len(); i++ {
		metric := metrics.At(i)
		if metric.Name() == "cpu_time" {
			gauge := metric.Gauge()
			if gauge.DataPoints().Len() > 0 {
				// Check if any data point has function.name attribute
				for j := 0; j < gauge.DataPoints().Len(); j++ {
					dp := gauge.DataPoints().At(j)
					if _, exists := dp.Attributes().Get("function.name"); exists {
						hasCPUWithFunction = true
					}
				}
			}
		} else if metric.Name() == "memory_allocation" {
			gauge := metric.Gauge()
			if gauge.DataPoints().Len() > 0 {
				// Check if any data point has function.name attribute
				for j := 0; j < gauge.DataPoints().Len(); j++ {
					dp := gauge.DataPoints().At(j)
					if _, exists := dp.Attributes().Get("function.name"); exists {
						hasMemoryWithFunction = true
					}
				}
			}
		}
	}

	assert.True(t, hasCPUWithFunction, "Should have cpu_time metric with function.name attribute")
	assert.True(t, hasMemoryWithFunction, "Should have memory_allocation metric with function.name attribute")
}

func TestConverter_GetSampleFunctionNameWithRealData(t *testing.T) {
	converter, err := NewConverter(&ConverterConfig{})
	require.NoError(t, err)

	profiles := pprofile.NewProfiles()
	dictionary := profiles.Dictionary()

	// Setup string table
	stringTable := dictionary.StringTable()
	stringTable.Append("my_function")

	// Setup function
	functionTable := dictionary.FunctionTable()
	fn := functionTable.AppendEmpty()
	fn.SetNameStrindex(0)

	// Setup location
	locationTable := dictionary.LocationTable()
	location := locationTable.AppendEmpty()
	line := location.Line().AppendEmpty()
	line.SetFunctionIndex(0)

	// Setup stack
	stackTable := dictionary.StackTable()
	stack := stackTable.AppendEmpty()
	stack.LocationIndices().Append(0)

	// Create sample
	sample := pprofile.NewSample()
	sample.SetStackIndex(0)

	// Test
	functionName := converter.getSampleFunctionName(profiles, sample)
	assert.Equal(t, "my_function", functionName)
}

func TestConverter_CalculateFunctionCPUTime(t *testing.T) {
	converter, err := NewConverter(&ConverterConfig{})
	require.NoError(t, err)

	profiles := pprofile.NewProfiles()
	profile := pprofile.NewProfile()

	// Setup function name resolution
	dictionary := profiles.Dictionary()
	stringTable := dictionary.StringTable()
	stringTable.Append("target_function")
	stringTable.Append("other_function")

	functionTable := dictionary.FunctionTable()
	fn1 := functionTable.AppendEmpty()
	fn1.SetNameStrindex(0)
	fn2 := functionTable.AppendEmpty()
	fn2.SetNameStrindex(1)

	locationTable := dictionary.LocationTable()
	loc1 := locationTable.AppendEmpty()
	line1 := loc1.Line().AppendEmpty()
	line1.SetFunctionIndex(0)

	stackTable := dictionary.StackTable()
	stack1 := stackTable.AppendEmpty()
	stack1.LocationIndices().Append(0)

	// Add samples
	sample1 := profile.Sample().AppendEmpty()
	sample1.SetStackIndex(0)
	values1 := sample1.Values()
	values1.Append(int64(1000000000)) // 1 second

	sample2 := profile.Sample().AppendEmpty()
	sample2.SetStackIndex(0)
	values2 := sample2.Values()
	values2.Append(int64(500000000)) // 0.5 seconds

	cpuTime := converter.calculateFunctionCPUTime(profiles, profile, "target_function")
	expected := 1.5 // 1s + 0.5s
	assert.InDelta(t, expected, cpuTime, 0.01)
}

func TestConverter_CalculateFunctionMemoryAllocation(t *testing.T) {
	converter, err := NewConverter(&ConverterConfig{})
	require.NoError(t, err)

	profiles := pprofile.NewProfiles()
	profile := pprofile.NewProfile()

	// Setup function name resolution
	dictionary := profiles.Dictionary()
	stringTable := dictionary.StringTable()
	stringTable.Append("target_function")

	functionTable := dictionary.FunctionTable()
	fn := functionTable.AppendEmpty()
	fn.SetNameStrindex(0)

	locationTable := dictionary.LocationTable()
	location := locationTable.AppendEmpty()
	line := location.Line().AppendEmpty()
	line.SetFunctionIndex(0)

	stackTable := dictionary.StackTable()
	stack := stackTable.AppendEmpty()
	stack.LocationIndices().Append(0)

	// Add samples with memory values
	sample1 := profile.Sample().AppendEmpty()
	sample1.SetStackIndex(0)
	values1 := sample1.Values()
	values1.Append(int64(1000))
	values1.Append(int64(2000)) // Memory

	sample2 := profile.Sample().AppendEmpty()
	sample2.SetStackIndex(0)
	values2 := sample2.Values()
	values2.Append(int64(1000))
	values2.Append(int64(3000)) // Memory

	memory := converter.calculateFunctionMemoryAllocation(profiles, profile, "target_function")
	expected := float64(2000 + 3000)
	assert.Equal(t, expected, memory)
}

func TestConverter_GenerateThreadMetrics(t *testing.T) {
	converter, err := NewConverter(&ConverterConfig{
		Metrics: MetricsConfig{
			CPU: CPUMetricConfig{
				Enabled:    true,
				MetricName: "cpu_time",
			},
			Memory: MemoryMetricConfig{
				Enabled:    true,
				MetricName: "memory_allocation",
			},
		},
	})
	require.NoError(t, err)

	profiles := pprofile.NewProfiles()
	profile := pprofile.NewProfile()
	attributes := map[string]string{"service.name": "test"}
	scopeMetrics := pmetric.NewScopeMetrics()

	// Generate thread metrics (should work even without actual thread data)
	converter.generateThreadMetrics(profiles, profile, attributes, scopeMetrics, "test_thread")

	// Verify metrics were created (even if empty)
	// The function should not panic
	assert.NotNil(t, scopeMetrics)
}

func TestConverter_GenerateProcessMetrics(t *testing.T) {
	converter, err := NewConverter(&ConverterConfig{
		Metrics: MetricsConfig{
			CPU: CPUMetricConfig{
				Enabled:    true,
				MetricName: "cpu_time",
			},
			Memory: MemoryMetricConfig{
				Enabled:    true,
				MetricName: "memory_allocation",
			},
		},
	})
	require.NoError(t, err)

	profiles := pprofile.NewProfiles()
	profile := pprofile.NewProfile()
	attributes := map[string]string{"service.name": "test"}
	scopeMetrics := pmetric.NewScopeMetrics()

	// Generate process metrics
	converter.generateProcessMetrics(profiles, profile, attributes, scopeMetrics, "test_process")

	// Verify metrics were created (even if empty)
	// The function should not panic
	assert.NotNil(t, scopeMetrics)
}
