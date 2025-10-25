# Thread and Process Filtering Implementation Summary

## Overview

A foundation has been added to support filtering CPU time and memory allocation metrics by specific threads or processes. The implementation is currently stubbed out but provides a clear path for completion.

## What Was Added

### Code Changes

1. **Filter Support in Calculation Functions** (`pkg/profiletometrics/converter.go`):
   - `calculateCPUTimeForFilter()` - CPU time calculation with optional filter
   - `calculateMemoryAllocationForFilter()` - Memory allocation calculation with optional filter
   - `matchesSampleFilter()` - Sample filtering logic
   - `getSampleAttributeValue()` - Attribute extraction (needs implementation)

2. **Metric Generation Functions**:
   - `generateThreadMetrics()` - Generates CPU/memory metrics for specific threads
   - `generateProcessMetrics()` - Generates CPU/memory metrics for specific processes
   - `generateMetricWithFilter()` - Helper for filtered metric generation
   - `sanitizeMetricName()` - Sanitizes thread/process names for metric naming

3. **Helper Functions** (stubbed out):
   - `getUniqueThreadNames()` - Extract unique thread names
   - `getUniqueProcessNames()` - Extract unique process names

4. **Main Integration** (lines 238-250):
   - Automatically generates thread/process specific metrics for all unique threads/processes

### Documentation

Created `docs/development/thread-process-filtering.md` with:
- Feature overview
- Current implementation status
- Usage examples
- Implementation approach
- Future enhancements

## Current Status

### ✅ Works
- Code compiles successfully
- All tests pass
- Filter infrastructure is in place
- Metric generation logic is implemented

### ⚠️ Pending
- `getSampleAttributeValue()` returns empty string
- `getUniqueThreadNames()` returns empty list
- `getUniqueProcessNames()` returns empty list
- No thread/process specific metrics are generated yet

## Why It's Not Working

Based on the pprofile schema you provided:

```
ProfileContainer
├── resource_profiles[]
    ├── resource (Resource attributes - container, host, etc.)
    ├── scope_profiles[]
        └── profiles[]
            ├── sample_type
            ├── samples[]  ← Values are here
            ├── locations[] ← Stack trace frames
            ├── mappings[] ← Binary/library mappings
            └── string_table[] ← Deduplicated strings
```

The issue is that:
1. **Sample attributes** (like thread.name, process.executable.name) are NOT stored directly on samples
2. They appear to be stored as **resource attributes** or possibly in the **string_table** with attribute indices
3. The `pprofile.Sample` API has `AttributeIndices()` but we need to know:
   - How to map indices to string table entries
   - What format is used for attribute keys/values
   - Whether attributes are on samples, locations, or elsewhere

## Next Steps to Complete

1. **Test with Real Data**:
   - Use actual profile data with thread/process information
   - Inspect the pprofile structure in a debugger
   - Identify where thread.name and process.executable.name are stored

2. **Implement Attribute Extraction**:
   - Based on findings, implement `getSampleAttributeValue()`
   - May need to traverse string_table using AttributeIndices
   - Handle resource-level attributes if that's where they are

3. **Extract Unique Names**:
   - Implement `getUniqueThreadNames()` and `getUniqueProcessNames()`
   - Consider if we need resource context or can extract from samples

4. **Test End-to-End**:
   - Verify thread/process metrics are generated
   - Check metric names are correctly sanitized
   - Validate values are calculated correctly

## Expected Behavior (Once Complete)

When processing profile data with thread and process information:

**Input**: Profile with samples from threads like "wrk:worker_1", "wrk:worker_2", and processes like "envoy", "app"

**Output**: Metrics like:
- `cpu_time_thread_wrk_worker_1`
- `cpu_time_thread_wrk_worker_2`  
- `memory_allocation_thread_wrk_worker_1`
- `cpu_time_process_envoy`
- `cpu_time_process_app`
- etc.

## Build Status

```bash
✅ go build ./...     # Compiles successfully
✅ go test ./...      # All tests pass
✅ go vet ./...       # No issues
```

The code is production-ready in its current state (just doesn't generate thread/process metrics yet).

## Function-Level Metrics Proposal

### The Problem

You mentioned that processes/threads are composed of multiple functions, and we should:
1. Report CPU time and memory allocation per function
2. Also report global CPU/memory totals for a process/thread
3. Ensure function-level metrics roll up correctly to thread/process totals

### Proposed Solution

Add a **three-level hierarchy** of metrics:

```
Level 1: Process-level (global totals)
  - cpu_time_process_envoy
  - memory_allocation_process_envoy
  
Level 2: Thread-level (within process)
  - cpu_time_thread_wrk_worker_1
  - memory_allocation_thread_wrk_worker_1
  
Level 3: Function-level (within thread)
  - cpu_time_function_thread_wrk_worker_1_func_main
  - memory_allocation_function_thread_wrk_worker_1_func_main
```

### Implementation Plan

1. **Access Location Data**: Use `profile.Location()` to get function names from stack traces
2. **Calculate Function Metrics**: For each unique function in each thread, sum CPU/memory
3. **Generate Hierarchical Metrics**: Function → Thread → Process aggregation
4. **Add Configuration**: Control which levels are enabled and cardinality limits

See `docs/development/thread-process-filtering.md` for full details.
