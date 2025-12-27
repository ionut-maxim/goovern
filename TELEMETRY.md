# OpenTelemetry Integration

This project includes full OpenTelemetry instrumentation for logs, metrics, and traces.

## Quick Start

### 1. Start the OTEL stack

```bash
docker-compose -f docker-compose.otel.yml up -d
```

This starts:
- **OTEL Collector** (`:4317` gRPC, `:4318` HTTP)
- **Jaeger** (`:16686` UI) - for traces
- **Prometheus** (`:9090`) - for metrics
- **Loki** (`:3100`) - for logs
- **Grafana** (`:3000`) - unified visualization

### 2. Enable telemetry in your app

```bash
export GOO_TELEMETRY_ENABLED=true
export GOO_TELEMETRY_OTEL_ENDPOINT=localhost:4317
export GOO_TELEMETRY_SERVICE_NAME=goovern
export GOO_TELEMETRY_SERVICE_VERSION=0.1.0
```

Or in your `.env`:
```
GOO_TELEMETRY_ENABLED=true
GOO_TELEMETRY_OTEL_ENDPOINT=localhost:4317
GOO_TELEMETRY_SERVICE_NAME=goovern
GOO_TELEMETRY_SERVICE_VERSION=0.1.0
```

### 3. Run your application

```bash
go run ./cmd/goovernd
```

## Accessing the UIs

- **Grafana**: http://localhost:3000 (unified dashboards)
- **Jaeger**: http://localhost:16686 (traces)
- **Prometheus**: http://localhost:9090 (metrics)

## What's Instrumented

### Logs
- All `slog` logs are automatically bridged to OTEL via `otelslog`
- Logs are sent to both console (stdout) for visibility AND OTEL collector via OTLP
- OTEL collector forwards logs to Loki with indexed labels for fast querying
- Logs within a trace span automatically include `trace_id`, `span_id`, and `trace_flags`
- Use `telemetry.LoggerWithTrace(ctx, logger)` to add trace context to logs
- **Indexed labels in Loki**: `service.name`, `service.version`, `deployment.environment`, `log.level`, `worker`

### Traces
- Example instrumentation in `DownloadWorker.Work()`
- Spans include resource IDs, names, job attempts, and priorities
- All errors are recorded on spans

### Metrics
- OTEL SDK auto-metrics (runtime, process info)
- Custom metrics can be added using `telemetry.Meter()`

## Adding Custom Instrumentation

### Traces with Log Correlation

```go
import "github.com/ionut-maxim/goovern/telemetry"
import "go.opentelemetry.io/otel/attribute"

func YourFunction(ctx context.Context, logger *slog.Logger) error {
    ctx, span := telemetry.StartSpan(ctx, "your.operation",
        attribute.String("key", "value"),
    )
    defer span.End()

    // Create logger with trace context for automatic correlation
    logger = telemetry.LoggerWithTrace(ctx, logger)

    logger.Info("Operation started")  // This log will have trace_id and span_id

    // Your code here

    if err != nil {
        logger.Error("Operation failed", "error", err)
        telemetry.RecordError(span, err)
        return err
    }

    logger.Info("Operation completed")
    return nil
}
```

### Metrics

```go
import "github.com/ionut-maxim/goovern/telemetry"

meter := telemetry.Meter()
counter, _ := meter.Int64Counter("goovern.downloads.total")
counter.Add(ctx, 1, metric.WithAttributes(
    attribute.String("resource.name", name),
))
```

## Architecture

```
┌─────────────┐
│   goovern   │
│ application │
└──────┬──────┘
       │ gRPC (4317)
       ▼
┌──────────────────┐
│ OTEL Collector   │
│  (processes &    │
│   routes data)   │
└───┬──────┬───┬───┘
    │      │   │
    ▼      ▼   ▼
┌────────┐ │ ┌──────┐
│ Jaeger │ │ │ Loki │
│(traces)│ │ │(logs)│
└────────┘ │ └──────┘
           ▼
    ┌─────────────┐
    │ Prometheus  │
    │  (metrics)  │
    └──────┬──────┘
           │
           ▼
    ┌─────────────┐
    │   Grafana   │
    │ (visualize) │
    └─────────────┘
```

## Configuration

All config is via environment variables with the `GOO_TELEMETRY_` prefix:

- `ENABLED` - Enable/disable OTEL (default: `false`)
- `OTEL_ENDPOINT` - OTEL collector endpoint (default: `localhost:4317`)
- `SERVICE_NAME` - Service name for telemetry (default: `goovern`)
- `SERVICE_VERSION` - Service version (default: `0.1.0`)

## Logging Behavior

When OTEL is **enabled**:
- Logs go to both stdout (JSON format) AND OTEL collector (via OTLP/gRPC)
- OTEL collector forwards to Loki with structured attributes
- Logs are automatically correlated with traces (trace_id/span_id embedded)
- View in Grafana's Explore with Loki datasource
- Stdout logs available for development/debugging

When OTEL is **disabled**:
- Normal slog behavior (stdout/stderr only)
- Uses your configured `GOO_LOG_TYPE` (pretty/json/text)
- No correlation with traces/metrics

## Trace-Log Correlation in Grafana

When viewing logs in Grafana's Explore:
1. Filter logs by service: `{service_name="goovern"}`
2. Logs within traces will have `trace_id` field
3. Click on a `trace_id` value to jump to the corresponding trace in Jaeger
4. In Jaeger, click on a span to see related logs

**Example Loki queries:**
```logql
# All logs for a specific trace
{service_name="goovern"} | json | trace_id="abc123..."

# All error logs with traces
{service_name="goovern", log_level="error"} | json | trace_id != ""

# Logs from download worker
{service_name="goovern", worker="download"}
```

## Production Considerations

1. **Sampling**: Update `sdktrace.WithSampler()` in `telemetry/otel.go`
2. **Batch sizes**: Adjust in `otel-collector-config.yaml`
3. **Retention**: Configure in Loki/Prometheus configs
4. **TLS**: Enable in production by removing `WithInsecure()`
5. **Resource limits**: See `memory_limiter` in collector config
