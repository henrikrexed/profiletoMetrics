# Testing Guide for Profile to Metrics Connector

This document describes the testing strategy and how to run tests for the Profile to Metrics Connector.

## Test Structure

The test suite is organized into several categories:

### 1. Unit Tests

#### Configuration Tests (`internal/config/config_test.go`)
- **Config Validation**: Tests for valid and invalid configuration scenarios
- **Default Values**: Ensures proper default values are set
- **Regex Compilation**: Validates regex pattern compilation
- **Attribute Extraction Rules**: Tests different extraction methods (literal, regex, string table index)

#### Connector Logic Tests (`internal/connector/connector_test.go`)
- **Profile Processing**: Tests profile data processing and metric generation
- **Attribute Extraction**: Tests attribute extraction from profiles
- **CPU Time Calculation**: Tests CPU time metric calculation
- **Memory Allocation Calculation**: Tests memory allocation metric calculation
- **Process Filtering**: Tests process name pattern filtering
- **Pattern Filtering**: Tests attribute pattern-based filtering

### 2. Integration Tests (`integration_test.go`)

#### End-to-End Testing
- **Basic Profile Processing**: Tests complete profile-to-metrics conversion
- **Java Application Profiles**: Tests with Java-specific profile data
- **Python Application Profiles**: Tests with Python-specific profile data
- **Process Filtering**: Tests filtering by process name patterns
- **Pattern Filtering**: Tests filtering by attribute patterns
- **String Table Extraction**: Tests regex-based extraction from string tables
- **Per-Process Metrics**: Tests generation of separate metrics per process

### 3. Test Data (`testdata/profile_test_data.go`)

#### Mock Profile Data
- **Basic Test Profile**: Simple profile with CPU and memory data
- **Java Application Profile**: Complex Java application profile with realistic function names
- **Python Application Profile**: Python application profile with data science libraries

## Running Tests

### Prerequisites

1. Ensure Go 1.22+ is installed
2. Install dependencies:
   ```bash
   go mod tidy
   ```

### Running All Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific test packages
go test -v ./internal/config/...
go test -v ./internal/connector/...
go test -v ./testdata/...
```

### Using the Test Runner Script

```bash
# Make the script executable
chmod +x run_tests.sh

# Run all tests with coverage
./run_tests.sh
```

### Running Individual Tests

```bash
# Run specific test functions
go test -v -run TestConfig_Validate ./internal/config/...
go test -v -run TestProfileToMetricsConnector_ConsumeProfiles ./internal/connector/...
go test -v -run TestIntegration_ProfileToMetricsConnector ./...
```

## Test Scenarios

### Configuration Tests

1. **Valid Configurations**
   - CPU time metrics enabled
   - Memory allocation metrics enabled
   - Both metrics enabled
   - Regex pattern extraction
   - Process filtering
   - Pattern filtering

2. **Invalid Configurations**
   - No metrics enabled
   - Invalid regex patterns
   - Missing required fields
   - Invalid extraction methods

### Connector Tests

1. **Metric Generation**
   - CPU time metrics with proper attributes
   - Memory allocation metrics with proper attributes
   - Both metrics generated simultaneously

2. **Attribute Extraction**
   - Literal value extraction
   - Regex pattern extraction from string table
   - String table index extraction
   - Resource attribute extraction

3. **Filtering**
   - Process name pattern filtering
   - Attribute pattern filtering
   - Combined filtering scenarios

### Integration Tests

1. **Real Profile Data**
   - Java application profiles with realistic function names
   - Python application profiles with data science libraries
   - Mixed workload scenarios

2. **Filtering Scenarios**
   - Production vs development namespace filtering
   - Java vs Python process filtering
   - Service name pattern filtering

3. **Per-Process Metrics**
   - Multiple processes generating separate metrics
   - Process-specific attribute extraction

## Test Data Examples

### Basic Profile Structure
```go
// CPU time: 1ms, 1.1ms, 1.2ms, 1.3ms, 1.4ms
// Memory: 1KB, 1.5KB, 2KB, 2.5KB, 3KB
// String table: ["main", "com.example.Main.main", ...]
```

### Java Application Profile
```go
// Higher CPU usage: 5ms to 9.5ms
// Higher memory allocation: 8KB to 17KB
// String table: ["com.example.api.UserController.getUser", ...]
```

### Python Application Profile
```go
// Moderate CPU usage: 2ms to 3.75ms
// Moderate memory allocation: 4KB to 7.5KB
// String table: ["pandas.DataFrame.read_csv", ...]
```

## Expected Test Results

### Unit Tests
- **Config Tests**: 15+ test cases covering validation, defaults, and regex compilation
- **Connector Tests**: 10+ test cases covering metric generation and filtering

### Integration Tests
- **Basic Profile**: 1 metric with CPU and memory data
- **Java Profile**: 1 metric with higher resource usage
- **Python Profile**: 1 metric with moderate resource usage
- **Filtered Scenarios**: 0 metrics (correctly filtered out)

### Coverage Goals
- **Overall Coverage**: >80%
- **Critical Paths**: >90%
- **Configuration Logic**: >95%
- **Metric Generation**: >85%

## Debugging Tests

### Verbose Output
```bash
go test -v -count=1 ./...
```

### Test-Specific Debugging
```bash
# Run single test with verbose output
go test -v -run TestConfig_Validate ./internal/config/...

# Run with race detection
go test -race ./...

# Run with memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof
```

### Common Issues

1. **Import Errors**: Ensure all dependencies are installed with `go mod tidy`
2. **Test Data Issues**: Verify test data structure matches expected profile format
3. **Mock Consumer Issues**: Check that mock consumer properly captures metrics
4. **Regex Compilation**: Ensure regex patterns are valid and properly escaped

## Continuous Integration

The test suite is designed to run in CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run Tests
  run: |
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
```

## Performance Testing

For performance testing with large profile datasets:

```bash
# Run benchmarks
go test -bench=. ./...

# Run with memory profiling
go test -bench=. -memprofile=mem.prof ./...
```

## Test Maintenance

1. **Adding New Tests**: Follow the existing pattern in test files
2. **Test Data Updates**: Update test data when profile schema changes
3. **Coverage Monitoring**: Maintain >80% test coverage
4. **Performance Regression**: Monitor test execution time

## Troubleshooting

### Common Test Failures

1. **Configuration Validation Failures**
   - Check regex patterns for validity
   - Verify required fields are set
   - Ensure extraction methods are supported

2. **Metric Generation Failures**
   - Verify profile data structure
   - Check attribute extraction rules
   - Validate metric names and descriptions

3. **Filtering Failures**
   - Verify regex patterns match expected data
   - Check attribute names and values
   - Ensure filtering logic is correct

### Debug Commands

```bash
# Run tests with detailed output
go test -v -count=1 ./...

# Run specific failing test
go test -v -run TestSpecificTest ./...

# Check test coverage
go test -cover ./...
go tool cover -func=coverage.out
```
