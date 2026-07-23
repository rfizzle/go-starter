# Application observability

This package owns process-level logger, tracer, and meter construction for the
application. It is internal platform code rather than a reusable telemetry
library.

`NewLogger` creates the configured Zap logger. `Init` installs OpenTelemetry
trace and metric providers backed by OTLP/gRPC exporters. Export is disabled by
default; the disabled path leaves the global no-op providers in place and does
not connect to a collector. The application composition root owns provider
shutdown so buffered signals can be flushed during graceful termination.

Business and domain packages must not construct exporters or loggers. Process
composition roots create them here and inject only the dependencies their
consumers require.

See [the environment guide](../../../docs/development/environment.md) for the
supported logging and OpenTelemetry settings.
