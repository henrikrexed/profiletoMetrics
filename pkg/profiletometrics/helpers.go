package profiletometrics

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pprofile"
)

// getSampleAttributeValueCommon returns the string value for a given attribute key in a sample.
func getSampleAttributeValueCommon(profiles pprofile.Profiles, sample pprofile.Sample, key string) string {
	attributeIndices := sample.AttributeIndices()
	if attributeIndices.Len() == 0 {
		return ""
	}

	dictionary := profiles.Dictionary()
	attributeTable := dictionary.AttributeTable()
	stringTable := dictionary.StringTable()

	for i := 0; i < attributeIndices.Len(); i++ {
		attrIndex := attributeIndices.At(i)
		if attrIndex < 0 || int(attrIndex) >= attributeTable.Len() {
			continue
		}

		attr := attributeTable.At(int(attrIndex))

		keyIndex := attr.KeyStrindex()
		if keyIndex < 0 || int(keyIndex) >= stringTable.Len() {
			continue
		}

		attrKey := stringTable.At(int(keyIndex))
		if attrKey == key {
			value := attr.Value()
			return value.AsString()
		}
	}

	return ""
}

// getLocationFileNameCommon returns the filename for the first line's function of a location.
func getLocationFileNameCommon(profiles pprofile.Profiles, location pprofile.Location) string {
	lines := location.Line()
	if lines.Len() == 0 {
		return ""
	}

	line := lines.At(0)
	functionIndex := line.FunctionIndex()
	if functionIndex < 0 {
		return ""
	}

	dictionary := profiles.Dictionary()
	functionTable := dictionary.FunctionTable()
	if int(functionIndex) >= functionTable.Len() {
		return ""
	}

	function := functionTable.At(int(functionIndex))
	filenameIndex := function.FilenameStrindex()

	stringTable := dictionary.StringTable()
	if filenameIndex < 0 || int(filenameIndex) >= stringTable.Len() {
		return ""
	}

	return stringTable.At(int(filenameIndex))
}

// getUniqueAttributeValuesCommon collects unique values of a sample attribute key across a profile.
func getUniqueAttributeValuesCommon(profiles pprofile.Profiles, profile pprofile.Profile, key string) []string {
	values := make(map[string]bool)
	for i := 0; i < profile.Sample().Len(); i++ {
		sample := profile.Sample().At(i)
		v := getSampleAttributeValueCommon(profiles, sample, key)
		if v != "" {
			values[v] = true
		}
	}
	var out []string
	for v := range values {
		out = append(out, v)
	}
	return out
}

// iterateProfilesCommon walks resource/scope/profile and calls back with extracted resource attributes
func iterateProfilesCommon(
	profiles pprofile.Profiles,
	extractResourceAttributes func(pcommon.Resource) map[string]string,
	onProfile func(resourceIndex, scopeIndex, profileIndex int, profile pprofile.Profile, resourceAttributes map[string]string),
) {
	for i := 0; i < profiles.ResourceProfiles().Len(); i++ {
		resourceProfile := profiles.ResourceProfiles().At(i)
		resourceAttributes := extractResourceAttributes(resourceProfile.Resource())
		for j := 0; j < resourceProfile.ScopeProfiles().Len(); j++ {
			scopeProfile := resourceProfile.ScopeProfiles().At(j)
			for k := 0; k < scopeProfile.Profiles().Len(); k++ {
				profile := scopeProfile.Profiles().At(k)
				onProfile(i, j, k, profile, resourceAttributes)
			}
		}
	}
}
