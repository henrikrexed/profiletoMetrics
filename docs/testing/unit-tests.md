# Unit Tests

This guide covers running and writing unit tests for the ProfileToMetrics Connector.

## Running Tests

### Basic Test Execution

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run tests with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Using Makefile

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Run specific test package
make test PKG=pkg/profiletometrics

# Run tests with race detection
make test-race
```

### Test Scripts

```bash
# Run all tests with coverage
./run_tests.sh

# Run specific test
go test -v ./pkg/profiletometrics -run TestConvertProfilesToMetrics
```

## Test Structure

### Package Organization

```
pkg/profiletometrics/
├── converter.go
├── converter_test.go
├── converter_connector.go
├── config.go
└── config_test.go
```

### Test Files

Each Go package should have corresponding test files:

- `converter_test.go` - Tests for the converter logic
- `config_test.go` - Tests for configuration validation
- `connector_test.go` - Tests for the connector implementation

## Writing Tests

### Basic Test Structure

```go
package profiletometrics

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestConvertProfilesToMetrics(t *testing.T) {
    // Arrange
    config := ConverterConfig{
        Metrics: MetricsConfig{
            CPU: CPUMetricsConfig{
                Enabled: true,
                MetricName: "cpu_time",
            },
            Memory: MemoryMetricsConfig{
                Enabled: true,
                MetricName: "memory_allocation",
            },
        },
    }
    
    // Act
    result, err := ConvertProfilesToMetrics(context.Background(), profiles)
    
    // Assert
    require.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, 2, result.ResourceMetrics().Len())
}
```

### Test Data

Use the provided test data in `testdata/profile_test_data.go`:

```go
func TestWithTestData(t *testing.T) {
    // Use test data
    profiles := testdata.CreateTestProfiles()
    
    // Test conversion
    result, err := ConvertProfilesToMetrics(context.Background(), profiles)
    
    require.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Mocking

For testing external dependencies:

```go
type MockConverter struct {
    ConvertProfilesToMetricsFunc func(context.Context, pprofile.Profiles) (pmetric.Metrics, error)
}

func (m *MockConverter) ConvertProfilesToMetrics(ctx context.Context, profiles pprofile.Profiles) (pmetric.Metrics, error) {
    if m.ConvertProfilesToMetricsFunc != nil {
        return m.ConvertProfilesToMetricsFunc(ctx, profiles)
    }
    return pmetric.NewMetrics(), nil
}

func TestConnectorWithMock(t *testing.T) {
    mockConverter := &MockConverter{}
    mockConverter.ConvertProfilesToMetricsFunc = func(ctx context.Context, profiles pprofile.Profiles) (pmetric.Metrics, error) {
        // Mock implementation
        return pmetric.NewMetrics(), nil
    }
    
    // Test with mock
    connector := &profileToMetricsConnector{
        converter: mockConverter,
    }
    
    // Test logic
}
```

## Test Categories

### Unit Tests

Test individual functions and methods:

```go
func TestCalculateCPUTime(t *testing.T) {
    tests := []struct {
        name     string
        samples  []pprofile.Sample
        expected float64
    }{
        {
            name:     "empty samples",
            samples:  []pprofile.Sample{},
            expected: 0.0,
        },
        {
            name:     "single sample",
            samples:  []pprofile.Sample{createTestSample()},
            expected: 0.123,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateCPUTime(tt.samples)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Integration Tests

Test the complete flow:

```go
func TestIntegration(t *testing.T) {
    // Create test profiles
    profiles := testdata.CreateTestProfiles()
    
    // Create converter
    converter := NewConverter(ConverterConfig{
        Metrics: MetricsConfig{
            CPU: CPUMetricsConfig{Enabled: true},
            Memory: MemoryMetricsConfig{Enabled: true},
        },
    })
    
    // Convert profiles
    result, err := converter.ConvertProfilesToMetrics(context.Background(), profiles)
    
    // Verify result
    require.NoError(t, err)
    assert.NotNil(t, result)
    
    // Check metrics
    resourceMetrics := result.ResourceMetrics()
    assert.Equal(t, 1, resourceMetrics.Len())
    
    scopeMetrics := resourceMetrics.At(0).ScopeMetrics()
    assert.Equal(t, 1, scopeMetrics.Len())
    
    metrics := scopeMetrics.At(0).Metrics()
    assert.Equal(t, 2, metrics.Len()) // CPU + Memory
}
```

### Configuration Tests

Test configuration validation:

```go
func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  MetricsConfig
        wantErr bool
    }{
        {
            name: "valid config",
            config: MetricsConfig{
                CPU: CPUMetricsConfig{Enabled: true},
                Memory: MemoryMetricsConfig{Enabled: true},
            },
            wantErr: false,
        },
        {
            name: "no metrics enabled",
            config: MetricsConfig{
                CPU: CPUMetricsConfig{Enabled: false},
                Memory: MemoryMetricsConfig{Enabled: false},
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Test Coverage

### Coverage Goals

- **Unit Tests**: 80%+ coverage
- **Integration Tests**: 60%+ coverage
- **Critical Paths**: 100% coverage

### Coverage Commands

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Coverage by function
go tool cover -func=coverage.out

# Coverage threshold check
go test -cover ./... | grep -E "coverage: [0-9]+\.[0-9]+%"
```

### Coverage Analysis

```bash
# Check coverage for specific package
go test -cover ./pkg/profiletometrics

# Coverage with race detection
go test -race -cover ./...

# Coverage for specific test
go test -cover -run TestConvertProfilesToMetrics ./pkg/profiletometrics
```

## Test Data

### Test Data Structure

```go
// testdata/profile_test_data.go
func CreateTestProfiles() pprofile.Profiles {
    profiles := pprofile.NewProfiles()
    resourceProfiles := profiles.ResourceProfiles()
    
    // Add resource profile
    resourceProfile := resourceProfiles.AppendEmpty()
    resourceProfile.Resource().Attributes().PutStr("service.name", "test-service")
    
    // Add scope profile
    scopeProfiles := resourceProfile.ScopeProfiles()
    scopeProfile := scopeProfiles.AppendEmpty()
    scopeProfile.Scope().SetName("test-scope")
    
    // Add profile
    profile := scopeProfile.Profiles().AppendEmpty()
    profile.SetStartTimeUnixNano(1000000000)
    profile.SetEndTimeUnixNano(2000000000)
    
    // Add samples
    samples := profile.Samples()
    sample := samples.AppendEmpty()
    sample.SetValue(0.123)
    
    return profiles
}
```

### Test Data Variants

```go
func CreateCPUTestProfiles() pprofile.Profiles {
    // CPU-specific test data
}

func CreateMemoryTestProfiles() pprofile.Profiles {
    // Memory-specific test data
}

func CreateFilteredTestProfiles() pprofile.Profiles {
    // Test data with filtering scenarios
}
```

## Performance Tests

### Benchmark Tests

```go
func BenchmarkConvertProfilesToMetrics(b *testing.B) {
    profiles := testdata.CreateTestProfiles()
    converter := NewConverter(ConverterConfig{
        Metrics: MetricsConfig{
            CPU: CPUMetricsConfig{Enabled: true},
            Memory: MemoryMetricsConfig{Enabled: true},
        },
    })
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := converter.ConvertProfilesToMetrics(context.Background(), profiles)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Memory Tests

```go
func TestMemoryUsage(t *testing.T) {
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Perform operation
    profiles := testdata.CreateTestProfiles()
    converter := NewConverter(ConverterConfig{})
    result, err := converter.ConvertProfilesToMetrics(context.Background(), profiles)
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    // Check memory usage
    memoryUsed := m2.Alloc - m1.Alloc
    assert.Less(t, memoryUsed, uint64(1024*1024)) // Less than 1MB
}
```

## Test Utilities

### Test Helpers

```go
func assertMetricsEqual(t *testing.T, expected, actual pmetric.Metrics) {
    t.Helper()
    
    expectedRM := expected.ResourceMetrics()
    actualRM := actual.ResourceMetrics()
    
    assert.Equal(t, expectedRM.Len(), actualRM.Len())
    
    for i := 0; i < expectedRM.Len(); i++ {
        expectedScope := expectedRM.At(i).ScopeMetrics()
        actualScope := actualRM.At(i).ScopeMetrics()
        
        assert.Equal(t, expectedScope.Len(), actualScope.Len())
    }
}

func createTestSample() pprofile.Sample {
    sample := pprofile.NewSample()
    sample.SetValue(0.123)
    return sample
}
```

### Test Fixtures

```go
func loadTestConfig(t *testing.T, filename string) ConverterConfig {
    t.Helper()
    
    data, err := os.ReadFile(filename)
    require.NoError(t, err)
    
    var config ConverterConfig
    err = yaml.Unmarshal(data, &config)
    require.NoError(t, err)
    
    return config
}
```

## Continuous Integration

### GitHub Actions

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.23'
      - name: Run tests
        run: make test
      - name: Run tests with coverage
        run: make test-coverage
```

### Test Reports

```bash
# Generate test report
go test -json ./... > test-results.json

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Troubleshooting Tests

### Common Issues

#### 1. Test Timeouts

```go
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Test with timeout
    result, err := ConvertProfilesToMetrics(ctx, profiles)
    require.NoError(t, err)
    assert.NotNil(t, result)
}
```

#### 2. Race Conditions

```bash
# Run tests with race detection
go test -race ./...

# Run specific test with race detection
go test -race -run TestConcurrentAccess ./pkg/profiletometrics
```

#### 3. Memory Leaks

```go
func TestMemoryLeak(t *testing.T) {
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Perform operation multiple times
    for i := 0; i < 1000; i++ {
        profiles := testdata.CreateTestProfiles()
        converter := NewConverter(ConverterConfig{})
        result, err := converter.ConvertProfilesToMetrics(context.Background(), profiles)
        require.NoError(t, err)
        assert.NotNil(t, result)
    }
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    // Check for memory leaks
    memoryGrowth := m2.Alloc - m1.Alloc
    assert.Less(t, memoryGrowth, uint64(1024*1024)) // Less than 1MB growth
}
```
