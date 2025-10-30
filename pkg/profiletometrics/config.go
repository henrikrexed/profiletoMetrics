package profiletometrics

// MetricsConfig defines the metrics configuration
type MetricsConfig struct {
	CPU      CPUMetricConfig      `mapstructure:"cpu"`
	Memory   MemoryMetricConfig   `mapstructure:"memory"`
	Function FunctionMetricConfig `mapstructure:"function"`
}

// CPUMetricConfig defines CPU metric configuration
type CPUMetricConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	MetricName string `mapstructure:"metric_name"`
	Unit       string `mapstructure:"unit"`
}

// MemoryMetricConfig defines memory metric configuration
type MemoryMetricConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	MetricName string `mapstructure:"metric_name"`
	Unit       string `mapstructure:"unit"`
}

// FunctionMetricConfig defines function-level metric configuration
type FunctionMetricConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// AttributeConfig defines attribute extraction configuration
type AttributeConfig struct {
	Key   string `mapstructure:"key"`
	Value string `mapstructure:"value"`
	Type  string `mapstructure:"type"` // "literal" or "regex"
}

// ProcessFilterConfig defines process filtering configuration
type ProcessFilterConfig struct {
	Enabled  bool     `mapstructure:"enabled"`
	Pattern  string   `mapstructure:"pattern"`  // backward-compat: single pattern
	Patterns []string `mapstructure:"patterns"` // preferred: list of patterns
}

// PatternFilterConfig defines pattern filtering configuration
type PatternFilterConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Pattern string `mapstructure:"pattern"`
}

// ThreadFilterConfig defines thread filtering configuration
type ThreadFilterConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Pattern string `mapstructure:"pattern"`
}
