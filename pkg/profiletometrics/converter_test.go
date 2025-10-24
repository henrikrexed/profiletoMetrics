package profiletometrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pprofile"

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
						Enabled: true,
						Name:    "test_cpu_time",
						Unit:    "s",
					},
					Memory: MemoryMetricConfig{
						Enabled: true,
						Name:    "test_memory_allocation",
						Unit:    "bytes",
					},
				},
				Attributes: []AttributeConfig{
					{
						Name:  "process_name",
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
						Enabled: true,
						Name:    "cpu_time_only",
						Unit:    "s",
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
						Enabled: true,
						Name:    "memory_only",
						Unit:    "bytes",
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

	// Add samples with CPU time values
	for i := 0; i < 3; i++ {
		sample := profile.Sample().AppendEmpty()
		values := sample.Values()
		values.Append(int64(1000 + i*100)) // 1000, 1100, 1200
	}

	cpuTime := converter.calculateCPUTime(profile)
	expected := float64(1000 + 1100 + 1200)
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

	// Add samples with memory allocation values (assuming second value is memory)
	for i := 0; i < 2; i++ {
		sample := profile.Sample().AppendEmpty()
		values := sample.Values()
		values.Append(1000)                // CPU time
		values.Append(int64(2000 + i*500)) // Memory: 2000, 2500
	}

	memoryAllocation := converter.calculateMemoryAllocation(profile)
	expected := float64(2000 + 2500)
	assert.Equal(t, expected, memoryAllocation)
}

func TestConverter_StringTableExtraction(t *testing.T) {
	config := &ConverterConfig{
		Metrics: MetricsConfig{
			CPU: CPUMetricConfig{
				Enabled: true,
				Name:    "cpu_time",
				Unit:    "s",
			},
		},
		Attributes: []AttributeConfig{
			{
				Name:  "function_name",
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
