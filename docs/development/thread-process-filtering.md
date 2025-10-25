# Thread and Process Filtering for Profile Metrics

## Overview

This feature enables users to calculate CPU time and memory allocation metrics for specific threads or processes within profile data. This allows for granular analysis of which threads or processes are consuming the most resources.

## Current Implementation Status

### ✅ Implemented

1. **Filter Support Infrastructure**:
   - Added `calculateCPUTimeForFilter()` method to calculate CPU time with optional filtering
   - Added `calculateMemoryAllocationForFilter()` method to calculate memory allocation with optional filtering
   - Added `matchesSampleFilter()` method to check if a sample matches filter criteria
   - Added helper functions `generateThreadMetrics()` and `generateProcessMetrics()` to generate metrics for specific threads/processes

2. **Metric Generation**:
   - The main profile processing loop calls `getUniqueThreadNames()` and `getUniqueProcessNames()` 
   - For each unique thread/process, it generates dedicated metrics with sanitized names
   - Metrics are named like `cpu_time_thread_thread_name` and `memory_allocation_process_process_name`

### ⚠️ Pending Implementation

The feature is currently stubbed out because the OpenTelemetry pprofile API doesn't provide direct access to sample attributes. The following functions need to be implemented based on the actual data structure:

1. **`getSampleAttributeValue()`**: Currently returns empty string. Needs to:
   - Parse attributes from the sample data structure
   - Extract thread.name, process.executable.name, or other attributes
   - Handle the actual pprofile data model (which may store attributes in location/mapping/labels)

2. **`getUniqueThreadNames()`**: Currently returns empty list. Needs to:
   - Iterate through profile samples
   - Extract unique thread names from sample attributes
   - Return a list of unique thread identifiers

3. **`getUniqueProcessNames()`**: Currently returns empty list. Needs to:
   - Iterate through profile samples  
   - Extract unique process names from sample attributes
   - Return a list of unique process identifiers

## How It Would Work (Once Implemented)

### Usage Example

When a user receives profile data with the following structure:

```
Sample #1:
  thread.name: "wrk:worker_1"
  process.executable.name: "envoy"
  
Sample #2:
  thread.name: "wrk:worker_2"
  process.executable.name: "envoy"
  
Sample #3:
  thread.name: "main"
  process.executable.name: "app"
```

The connector would automatically generate:

**Thread-Specific Metrics:**
- `cpu_time_thread_wrk_worker_1`
- `cpu_time_thread_wrk_worker_2`
- `cpu_time_thread_main`
- `memory_allocation_thread_wrk_worker_1`
- `memory_allocation_thread_wrk_worker_2`
- `memory_allocation_thread_main`

**Process-Specific Metrics:**
- `cpu_time_process_envoy`
- `cpu_time_process_app`
- `memory_allocation_process_envoy`
- `memory_allocation_process_app`

## Implementation Approach

To properly implement this feature, you need to:

1. **Understand the pprofile Data Structure**:
   - Research how OpenTelemetry stores sample-level attributes
   - Check if attributes are stored in Location, Mapping, or Label structures
   - Review the pprofile protobuf definitions

2. **Access Sample Attributes**:
   - Implement attribute extraction based on the actual data structure
   - May need to traverse profile.Location, profile.Mapping, or profile.Label structures
   - Handle different attribute storage mechanisms (string table, direct values, etc.)

3. **Handle Attribute Names**:
   - Standardize attribute key names (e.g., "thread.name" vs "thread_name")
   - Support multiple attribute naming conventions
   - Handle missing attributes gracefully

4. **Test with Real Data**:
   - Use actual profile data from eBPF profilers or similar tools
   - Verify attribute extraction works correctly
   - Test with various attribute formats and structures

## Code Locations

- Filter functions: `pkg/profiletometrics/converter.go` lines 304-421
- Helper functions: `pkg/profiletometrics/converter.go` lines 417-453  
- Main integration: `pkg/profiletometrics/converter.go` lines 238-250

## Function-Level Metrics (Proposed)

### Overview

Since processes and threads are composed of multiple functions, we should support:

1. **Function-level metrics**: CPU time and memory allocation per function
2. **Hierarchical aggregation**: Function metrics roll up to process/thread totals
3. **Global process totals**: Total CPU/memory across all threads in a process

### Proposed Metrics Structure

```
# Process-level metrics
cpu_time_process_envoy = <total CPU across all threads>
memory_allocation_process_envoy = <total memory across all threads>

# Thread-level metrics  
cpu_time_thread_wrk_worker_1 = <total CPU for this thread>
memory_allocation_thread_wrk_worker_1 = <total memory for this thread>

# Function-level metrics (within a thread)
cpu_time_function_main_thread_main_function = <CPU for main function in main thread>
cpu_time_function_main_thread_envoy_accept = <CPU for accept function in main thread>
memory_allocation_function_main_thread_main_function = <memory for main function>
```

### Implementation Approach

1. **Access Stack Traces**: Use `profile.Location()` to get function information
2. **Calculate Function Metrics**: Sum CPU/memory for each unique function
3. **Maintain Hierarchical Totals**: Ensure function totals equal thread totals, thread totals equal process totals
4. **Handle Cardinality**: Consider limits on function metrics to prevent explosion

### Configuration

```yaml
connectors:
  profiletometrics:
    metrics:
      cpu:
        enabled: true
      memory:
        enabled: true
      # Enable function-level metrics
      function_metrics:
        enabled: true
        max_functions: 100  # Limit to prevent cardinality explosion
    # Enable thread/process aggregation
    thread_metrics:
      enabled: true
    process_metrics:
      enabled: true
```

## Future Enhancements

Once the basic filtering is implemented, consider:

1. **Configuration Support**: Allow users to enable/disable thread/process/function metrics
2. **Pattern Matching**: Support regex patterns for thread/process/function names
3. **Aggregation**: Support min/max/avg aggregations per thread/process/function
4. **Cardinality Control**: Limit the number of unique threads/processes/functions to track
5. **Sampling**: Sample high-cardinality function data
6. **Call Tree Metrics**: Track parent-child relationships in call stacks
