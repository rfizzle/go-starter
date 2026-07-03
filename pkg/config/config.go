package config

type Application struct {
	Name string `mapstructure:"name"`
	// Environment is the runtime deployment environment (e.g. "development",
	// "staging", "production"). It selects logger defaults and is attached as a
	// field to every log line. It is distinct from Build.Env, which records the
	// environment the binary was built for.
	Environment string    `mapstructure:"environment"`
	Build       Build     `mapstructure:"build"`
	Webserver   Webserver `mapstructure:"webserver"`
	Database    Database  `mapstructure:"database"`
	Auth        Auth      `mapstructure:"auth"`
	Logging     Logging   `mapstructure:"logging"`
	// ShutdownDelay is the delay, in seconds, between receiving a shutdown
	// signal and beginning graceful shutdown. It gives orchestrators such as
	// Kubernetes time to propagate the not-ready state and stop routing traffic
	// before in-flight components are torn down. Default 0 (no delay).
	ShutdownDelay int `mapstructure:"shutdown-delay"`
}

type Build struct {
	Branch   string `mapstructure:"branch"`
	Date     string `mapstructure:"date"`
	Env      string `mapstructure:"env"`
	Revision string `mapstructure:"revision"`
	Version  string `mapstructure:"version"`
}

type Webserver struct {
	Address      string `mapstructure:"address"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read-timeout"`
	WriteTimeout int    `mapstructure:"write-timeout"`
	IdleTimeout  int    `mapstructure:"idle-timeout"`
}

type Database struct {
	// URI is the PostgreSQL connection string, e.g.
	// "postgres://user:password@localhost:5432/dbname".
	URI string `mapstructure:"uri"`
	// PingTimeout is the timeout, in seconds, for database ping operations.
	PingTimeout int `mapstructure:"ping-timeout"`
}

type Auth struct {
	// JWTSecret is the shared HMAC (HS256) secret used to sign and verify JWT
	// bearer tokens. When empty, protected endpoints fail closed (unauthorized).
	JWTSecret string `mapstructure:"jwt-secret"`
	// JWTIssuer, when set, is required on validated tokens ("iss" claim).
	JWTIssuer string `mapstructure:"jwt-issuer"`
	// JWTAudience, when set, is required on validated tokens ("aud" claim).
	JWTAudience string `mapstructure:"jwt-audience"`
	// JWTTTL is the lifetime, in seconds, of generated tokens. Default 3600.
	JWTTTL int `mapstructure:"jwt-ttl"`
}

type Logging struct {
	// Level is the minimum enabled log level: "debug", "info", "warn", or
	// "error". Empty defaults to "info".
	Level string `mapstructure:"level"`
	// Format is the log encoding: "json" or "console". Empty derives from the
	// environment (console in development, JSON otherwise).
	Format string `mapstructure:"format"`
}
