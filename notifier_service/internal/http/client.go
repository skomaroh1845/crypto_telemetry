package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Client is a wrapper around http.Client with OpenTelemetry instrumentation
type Client struct {
	client *http.Client
	tracer trace.Tracer
}

// NewClient creates a new HTTP client with OpenTelemetry support
func NewClient(timeout int) *Client {
	return &Client{
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		tracer: otel.Tracer("http-client"),
	}
}

// RequestOptions contains options for HTTP requests
type RequestOptions struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    io.Reader
}

// Do performs an HTTP request with OpenTelemetry instrumentation
func (c *Client) Do(ctx context.Context, opts RequestOptions, v interface{}) error {
	ctx, span := c.tracer.Start(ctx, "HTTP."+opts.Method)
	defer span.End()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, opts.Method, opts.URL, opts.Body)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}

	// Inject OpenTelemetry trace context
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Add tracing attributes
	span.SetAttributes(
		attribute.String("http.url", opts.URL),
		attribute.String("http.method", opts.Method),
	)

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Record response attributes
	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	// Check for HTTP error status
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
		span.RecordError(err)
		return err
	}

	// Decode response if target is provided
	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Get performs an HTTP GET request
func (c *Client) Get(ctx context.Context, url string, headers map[string]string, v interface{}) error {
	return c.Do(ctx, RequestOptions{
		Method:  "GET",
		URL:     url,
		Headers: headers,
	}, v)
}

// Post performs an HTTP POST request
func (c *Client) Post(ctx context.Context, url string, headers map[string]string, body io.Reader, v interface{}) error {
	return c.Do(ctx, RequestOptions{
		Method:  "POST",
		URL:     url,
		Headers: headers,
		Body:    body,
	}, v)
}
