package testdata

import (
	"go.opentelemetry.io/collector/pdata/pprofile"
)

// CreateTestProfile creates a test profile with sample data
func CreateTestProfile() pprofile.Profiles {
	profiles := pprofile.NewProfiles()
	resourceProfile := profiles.ResourceProfiles().AppendEmpty()

	// Add resource attributes
	resource := resourceProfile.Resource()
	resource.Attributes().PutStr("process.name", "test_application")
	resource.Attributes().PutStr("k8s.pod.name", "test-pod-123")
	resource.Attributes().PutStr("k8s.namespace.name", "default")
	resource.Attributes().PutStr("service.name", "test-service")

	// Add scope profile
	scopeProfile := resourceProfile.ScopeProfiles().AppendEmpty()
	scopeProfile.Scope().SetName("test-scope")
	scopeProfile.Scope().SetVersion("1.0.0")

	// Add profile
	profile := scopeProfile.Profiles().AppendEmpty()

	// Add string table with function names and sample type information
	stringTable := profile.StringTable()
	stringTable.Append("main")
	stringTable.Append("com.example.Main.main")
	stringTable.Append("com.example.Main.processRequest")
	stringTable.Append("com.example.Service.handleRequest")
	stringTable.Append("java.lang.String.toString")
	// Sample type strings
	stringTable.Append("cpu")         // index 5
	stringTable.Append("nanoseconds") // index 6
	stringTable.Append("alloc_space") // index 7
	stringTable.Append("bytes")       // index 8

	// Add sample types (CPU time and memory allocation)
	sampleTypes := profile.SampleType()

	// CPU time sample type
	cpuType := sampleTypes.AppendEmpty()
	cpuType.SetTypeStrindex(5) // "cpu"
	cpuType.SetUnitStrindex(6) // "nanoseconds"

	// Memory allocation sample type
	memType := sampleTypes.AppendEmpty()
	memType.SetTypeStrindex(7) // "alloc_space"
	memType.SetUnitStrindex(8) // "bytes"

	// Add attribute table with thread and process information
	attributeTable := profile.AttributeTable()

	// Thread attributes
	threadAttr1 := attributeTable.AppendEmpty()
	threadAttr1.SetKey("thread_name")
	threadAttr1.Value().SetStr("main-thread")

	threadAttr2 := attributeTable.AppendEmpty()
	threadAttr2.SetKey("thread_name")
	threadAttr2.Value().SetStr("worker-thread-1")

	threadAttr3 := attributeTable.AppendEmpty()
	threadAttr3.SetKey("thread_name")
	threadAttr3.Value().SetStr("worker-thread-2")

	// Process attributes
	processAttr1 := attributeTable.AppendEmpty()
	processAttr1.SetKey("process_name")
	processAttr1.Value().SetStr("test_application")

	processAttr2 := attributeTable.AppendEmpty()
	processAttr2.SetKey("process_name")
	processAttr2.Value().SetStr("background_service")

	// Add samples with CPU time and memory allocation
	for i := 0; i < 5; i++ {
		sample := profile.Sample().AppendEmpty()

		// CPU time value (in nanoseconds)
		sample.Value().Append(int64(1000000 + i*100000)) // 1ms, 1.1ms, 1.2ms, 1.3ms, 1.4ms

		// Memory allocation value (in bytes)
		sample.Value().Append(int64(1024 + i*512)) // 1KB, 1.5KB, 2KB, 2.5KB, 3KB

		// Add thread and process attributes to sample
		attributeIndices := sample.AttributeIndices()
		if i < 2 {
			// First two samples belong to main thread
			attributeIndices.Append(0) // thread_name: "main-thread"
			attributeIndices.Append(4) // process_name: "test_application"
		} else if i < 4 {
			// Next two samples belong to worker-thread-1
			attributeIndices.Append(1) // thread_name: "worker-thread-1"
			attributeIndices.Append(4) // process_name: "test_application"
		} else {
			// Last sample belongs to worker-thread-2
			attributeIndices.Append(2) // thread_name: "worker-thread-2"
			attributeIndices.Append(5) // process_name: "background_service"
		}
	}

	// Add locations (simplified for testing)
	// Note: Location and Function APIs might not be available in this version

	return profiles
}

// CreateJavaProfile creates a Java application profile
func CreateJavaProfile() pprofile.Profiles {
	profiles := pprofile.NewProfiles()
	resourceProfile := profiles.ResourceProfiles().AppendEmpty()

	// Add resource attributes for Java application
	resource := resourceProfile.Resource()
	resource.Attributes().PutStr("process.name", "java_application")
	resource.Attributes().PutStr("k8s.pod.name", "java-pod-prod-456")
	resource.Attributes().PutStr("k8s.namespace.name", "production")
	resource.Attributes().PutStr("service.name", "api-service")
	resource.Attributes().PutStr("runtime.name", "java")
	resource.Attributes().PutStr("runtime.version", "11.0.16")

	// Add scope profile
	scopeProfile := resourceProfile.ScopeProfiles().AppendEmpty()
	scopeProfile.Scope().SetName("java-profiler")
	scopeProfile.Scope().SetVersion("1.0.0")

	// Add profile
	profile := scopeProfile.Profiles().AppendEmpty()

	// Add string table with Java-specific function names and sample type information
	stringTable := profile.StringTable()
	stringTable.Append("main")
	stringTable.Append("com.example.api.UserController.getUser")
	stringTable.Append("com.example.api.UserController.createUser")
	stringTable.Append("com.example.service.UserService.findById")
	stringTable.Append("com.example.service.UserService.save")
	stringTable.Append("org.springframework.web.servlet.DispatcherServlet.doDispatch")
	stringTable.Append("org.springframework.web.servlet.DispatcherServlet.doService")
	stringTable.Append("java.util.HashMap.get")
	stringTable.Append("java.util.HashMap.put")
	stringTable.Append("java.lang.String.hashCode")
	// Sample type strings
	stringTable.Append("cpu")         // index 10
	stringTable.Append("nanoseconds") // index 11
	stringTable.Append("alloc_space") // index 12
	stringTable.Append("bytes")       // index 13

	// Add sample types
	sampleTypes := profile.SampleType()

	// CPU time sample type
	cpuType := sampleTypes.AppendEmpty()
	cpuType.SetTypeStrindex(10) // "cpu"
	cpuType.SetUnitStrindex(11) // "nanoseconds"

	// Memory allocation sample type
	memType := sampleTypes.AppendEmpty()
	memType.SetTypeStrindex(12) // "alloc_space"
	memType.SetUnitStrindex(13) // "bytes"

	// Add samples with higher CPU usage and memory allocation
	for i := 0; i < 10; i++ {
		sample := profile.Sample().AppendEmpty()

		// Higher CPU time for Java application
		sample.Value().Append(int64(5000000 + i*500000)) // 5ms to 9.5ms

		// Higher memory allocation
		sample.Value().Append(int64(8192 + i*1024)) // 8KB to 17KB
	}

	// Add locations and functions (simplified for testing)
	// Note: Location and Function APIs might not be available in this version

	return profiles
}

// CreatePythonProfile creates a Python application profile
func CreatePythonProfile() pprofile.Profiles {
	profiles := pprofile.NewProfiles()
	resourceProfile := profiles.ResourceProfiles().AppendEmpty()

	// Add resource attributes for Python application
	resource := resourceProfile.Resource()
	resource.Attributes().PutStr("process.name", "python_application")
	resource.Attributes().PutStr("k8s.pod.name", "python-pod-dev-789")
	resource.Attributes().PutStr("k8s.namespace.name", "development")
	resource.Attributes().PutStr("service.name", "data-service")
	resource.Attributes().PutStr("runtime.name", "python")
	resource.Attributes().PutStr("runtime.version", "3.9.7")

	// Add scope profile
	scopeProfile := resourceProfile.ScopeProfiles().AppendEmpty()
	scopeProfile.Scope().SetName("python-profiler")
	scopeProfile.Scope().SetVersion("1.0.0")

	// Add profile
	profile := scopeProfile.Profiles().AppendEmpty()

	// Add string table with Python-specific function names and sample type information
	stringTable := profile.StringTable()
	stringTable.Append("main")
	stringTable.Append("app.main")
	stringTable.Append("app.process_data")
	stringTable.Append("app.analyze_data")
	stringTable.Append("pandas.DataFrame.read_csv")
	stringTable.Append("pandas.DataFrame.groupby")
	stringTable.Append("numpy.array.sum")
	stringTable.Append("numpy.array.mean")
	stringTable.Append("sklearn.model_selection.train_test_split")
	stringTable.Append("sklearn.ensemble.RandomForestClassifier.fit")
	// Sample type strings
	stringTable.Append("cpu")         // index 10
	stringTable.Append("nanoseconds") // index 11
	stringTable.Append("alloc_space") // index 12
	stringTable.Append("bytes")       // index 13

	// Add sample types
	sampleTypes := profile.SampleType()

	// CPU time sample type
	cpuType := sampleTypes.AppendEmpty()
	cpuType.SetTypeStrindex(10) // "cpu"
	cpuType.SetUnitStrindex(11) // "nanoseconds"

	// Memory allocation sample type
	memType := sampleTypes.AppendEmpty()
	memType.SetTypeStrindex(12) // "alloc_space"
	memType.SetUnitStrindex(13) // "bytes"

	// Add samples with moderate CPU usage and memory allocation
	for i := 0; i < 8; i++ {
		sample := profile.Sample().AppendEmpty()

		// Moderate CPU time for Python application
		sample.Value().Append(int64(2000000 + i*250000)) // 2ms to 3.75ms

		// Moderate memory allocation
		sample.Value().Append(int64(4096 + i*512)) // 4KB to 7.5KB
	}

	// Add locations and functions (simplified for testing)
	// Note: Location and Function APIs might not be available in this version

	return profiles
}
