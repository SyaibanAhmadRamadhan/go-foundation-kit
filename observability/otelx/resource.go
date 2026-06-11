package otelx

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// NewOpenTelemetryBasicResource builds and returns a resource describing this service with only the service name.
// Its objective is to provide a quick and standard way to identify a service in traces and metrics
// without needing additional custom metadata.
func NewOpenTelemetryBasicResource(ctx context.Context, serviceName string) (*resource.Resource, error) {
	return NewOpenTelemetryResource(ctx, serviceName, "", "", "", "")
}

// NewOpenTelemetryResource builds and returns a resource describing this service with standard and custom attributes.
// Its objective is to ensure that essential metadata is always present for better filtering and analysis:
// - service.name: The application name.
// - service.namespace: An optional namespace for the service name.
// - deployment.environment: Name of the deployment environment (e.g., staging, production).
// - service.instance.id: The unique instance ID (e.g., pod name).
// - service.version: The application version.
func NewOpenTelemetryResource(ctx context.Context, serviceName, serviceNamespace, deploymentEnv, serviceInstanceID, serviceVersion string, attrs ...attribute.KeyValue) (*resource.Resource, error) {
	baseAttrs := []attribute.KeyValue{
		semconv.ServiceName(serviceName),
	}

	if serviceNamespace != "" {
		baseAttrs = append(baseAttrs, semconv.ServiceNamespace(serviceNamespace))
	}
	if deploymentEnv != "" {
		baseAttrs = append(baseAttrs, semconv.DeploymentEnvironment(deploymentEnv))
	}
	if serviceInstanceID != "" {
		baseAttrs = append(baseAttrs, semconv.ServiceInstanceID(serviceInstanceID))
	}
	if serviceVersion != "" {
		baseAttrs = append(baseAttrs, semconv.ServiceVersion(serviceVersion))
	}

	return resource.New(ctx,
		resource.WithAttributes(
			append(baseAttrs, attrs...)...,
		),
		resource.WithHost(),
		resource.WithTelemetrySDK(),
		resource.WithFromEnv(),
	)
}
