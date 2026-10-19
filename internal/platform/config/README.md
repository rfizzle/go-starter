# Application configuration

This package defines the application's decoded configuration values and
command-scoped validation rules. It is internal platform code because its
contract belongs to this binary rather than to external Go consumers.

The package is a side-effect-free leaf. It does not read files, environment
variables, flags, secrets, or remote configuration. Cobra and Viper wiring
lives in `cmd`, which resolves configuration into `Application` and validates
the command-specific subset before constructing network or database resources.

`Application.Validate` reports all detected problems together and never
includes configured values in its errors because some fields contain secrets.
Add a validation scope only when a command consumes a genuinely different
subset of settings.

Configuration precedence is in
[the architecture document](../../../docs/architecture.md#configuration-and-observability);
`.env.example` lists the environment variables, and each command's `--help`
lists every setting.
