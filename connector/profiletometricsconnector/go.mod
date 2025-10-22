module github.com/henrikrexed/profiletoMetrics/connector/profiletometricsconnector

go 1.24.0

require (
	github.com/henrikrexed/profiletoMetrics v0.1.0
	go.opentelemetry.io/collector/component v0.118.0
	go.opentelemetry.io/collector/connector v0.118.0
	go.opentelemetry.io/collector/consumer v1.24.0
	go.opentelemetry.io/collector/pdata v1.44.0
	go.uber.org/zap v1.27.0
)

replace github.com/henrikrexed/profiletoMetrics => /app
