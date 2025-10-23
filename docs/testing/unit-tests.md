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

## Test Categories

### Unit Tests

Test individual functions and methods:

- **Converter Logic**: Test profile-to-metrics conversion
- **Configuration Validation**: Test configuration parsing and validation
- **Filtering Logic**: Test process, thread, and pattern filters
- **Attribute Extraction**: Test string table extraction

### Integration Tests

Test the complete flow:

- **End-to-End Conversion**: Test full profile-to-metrics pipeline
- **Configuration Loading**: Test configuration file loading
- **Error Handling**: Test error scenarios and recovery
- **Performance**: Test with realistic data volumes

### Configuration Tests

Test configuration validation:

- **Valid Configurations**: Test various valid configuration combinations
- **Invalid Configurations**: Test error handling for invalid configs
- **Default Values**: Test default configuration behavior
- **Edge Cases**: Test boundary conditions

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

The connector includes comprehensive test data for:

- **CPU Profiling Data**: CPU time samples and call stacks
- **Memory Profiling Data**: Memory allocation samples
- **String Table Data**: Various string table configurations
- **Attribute Data**: Different attribute extraction scenarios

### Test Data Variants

- **CPU Test Data**: CPU-specific profiling samples
- **Memory Test Data**: Memory allocation samples
- **Filtered Test Data**: Test data with various filtering scenarios
- **Edge Case Data**: Boundary conditions and error cases

## Performance Tests

### Benchmark Tests

The connector includes benchmark tests for:

- **Conversion Performance**: Profile-to-metrics conversion speed
- **Memory Usage**: Memory consumption during processing
- **Throughput**: Processing rate with large datasets
- **Scalability**: Performance with increasing data volumes

### Memory Tests

Test memory usage and potential leaks:

- **Memory Allocation**: Track memory usage during processing
- **Memory Leaks**: Detect memory leaks in long-running operations
- **Garbage Collection**: Test GC behavior with large datasets
- **Resource Cleanup**: Verify proper resource cleanup

## Test Utilities

### Test Helpers

The connector provides test utilities for:

- **Metric Comparison**: Compare generated metrics with expected results
- **Configuration Loading**: Load test configurations from files
- **Mock Data**: Generate mock profiling data for testing
- **Assertion Helpers**: Custom assertions for OpenTelemetry data types

### Test Fixtures

- **Configuration Files**: Pre-defined test configurations
- **Sample Data**: Realistic profiling data samples
- **Expected Results**: Expected output for validation
- **Test Scenarios**: Common test scenarios and edge cases

## Continuous Integration

### GitHub Actions

The connector includes CI/CD pipelines for:

- **Automated Testing**: Run tests on every commit
- **Coverage Reporting**: Generate and report test coverage
- **Performance Testing**: Run performance benchmarks
- **Cross-Platform Testing**: Test on multiple operating systems

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

- **Long-Running Tests**: Use timeouts for long-running operations
- **Resource Cleanup**: Ensure proper cleanup in test teardown
- **Context Cancellation**: Use context for test cancellation

#### 2. Race Conditions

```bash
# Run tests with race detection
go test -race ./...

# Run specific test with race detection
go test -race -run TestConcurrentAccess ./pkg/profiletometrics
```

#### 3. Memory Leaks

- **Memory Monitoring**: Track memory usage in tests
- **Resource Cleanup**: Verify proper resource cleanup
- **Long-Running Tests**: Test memory behavior over time
- **Garbage Collection**: Test GC behavior with large datasets

### Debug Commands

```bash
# Run tests with debug output
go test -v ./...

# Run specific test with debug
go test -v -run TestSpecificFunction ./pkg/profiletometrics

# Run tests with race detection
go test -race ./...

# Run tests with memory profiling
go test -memprofile=mem.prof ./...
```

## Best Practices

### Test Organization

- **Package Structure**: Organize tests by package and functionality
- **Test Naming**: Use descriptive test names that explain the scenario
- **Test Data**: Use realistic test data that matches production scenarios
- **Test Isolation**: Ensure tests don't depend on each other

### Test Writing

- **Arrange-Act-Assert**: Follow the AAA pattern for test structure
- **Test Coverage**: Aim for high coverage of critical paths
- **Edge Cases**: Test boundary conditions and error scenarios
- **Performance**: Include performance tests for critical operations

### Test Maintenance

- **Regular Updates**: Keep tests updated with code changes
- **Test Documentation**: Document complex test scenarios
- **Test Review**: Review tests as part of code review process
- **Test Metrics**: Track test coverage and performance metrics