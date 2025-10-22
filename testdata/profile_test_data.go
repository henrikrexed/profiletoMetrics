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

	// Add samples with CPU time and memory allocation
	for i := 0; i < 5; i++ {
		sample := profile.Sample().AppendEmpty()

		// Add values to the sample
		values := sample.Values()
		values.Append(int64(1000000 + i*100000)) // CPU time in nanoseconds
		values.Append(int64(1024 + i*512))       // Memory allocation in bytes
	}

	return profiles
}
