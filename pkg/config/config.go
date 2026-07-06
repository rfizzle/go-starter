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
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
	// ReadHeaderTimeout is the timeout, in seconds, for reading request
	// headers. It bounds how long a client may take to send its headers,
	// mitigating Slowloris-style attacks that hold connections open by
	// trickling headers. 0 keeps this server's safe 5s default rather than
	// Go's own behaviour (fall back to ReadTimeout, else no timeout at all).
	ReadHeaderTimeout int `mapstructure:"read-header-timeout"`
	ReadTimeout       int `mapstructure:"read-timeout"`
	WriteTimeout      int `mapstructure:"write-timeout"`
	IdleTimeout       int `mapstructure:"idle-timeout"`
	// MaxHeaderBytes caps the total size, in bytes, of request headers the
	// server will read. It bounds the memory a single request can force the
	// server to buffer, mitigating oversized-header attacks. 0 keeps this
	// server's safe 1 MiB default.
	MaxHeaderBytes int `mapstructure:"max-header-bytes"`
	// MaxBodyBytes caps the size, in bytes, of each request body the server will
	// read. It bounds the memory a single POST/PUT can force the server to
	// buffer; reading past it fails the request (413 or 400 depending on the
	// handler). 0 or less disables the limit (unbounded bodies). Default 10 MiB.
	MaxBodyBytes int64 `mapstructure:"max-body-bytes"`
	// PprofEnabled mounts the /debug/pprof/* profiling endpoints on the public
	// mux. Disabled by default: the endpoints are unauthenticated and expose
	// process internals (an information-leak and DoS surface). Enable only in
	// trusted environments or behind separate access controls.
	PprofEnabled bool `mapstructure:"pprof-enabled"`
	// CORSAllowedOrigins is the list of origins permitted by the CORS
	// middleware's Access-Control-Allow-Origin handling. Each entry may contain
	// a "*" wildcard (e.g. "https://*.example.com"). The default of
	// "https://*"/"http://*" allows any origin and is a permissive placeholder:
	// restrict it to your front-end origins before deploying to production.
	CORSAllowedOrigins []string `mapstructure:"cors-allowed-origins"`
	// TLSCertFile and TLSKeyFile are paths to a PEM-encoded certificate and
	// private key. When both are set the server terminates TLS in-process
	// (negotiating a TLS 1.2 floor); when either is empty the server serves
	// plain HTTP and assumes TLS is terminated upstream by a reverse proxy or
	// load balancer — the default posture for containerised deployments.
	TLSCertFile string `mapstructure:"tls-cert-file"`
	TLSKeyFile  string `mapstructure:"tls-key-file"`
}

type Database struct {
	// URI is the PostgreSQL connection string, e.g.
	// "postgres://user:password@localhost:5432/dbname".
	URI string `mapstructure:"uri"`
	// PingTimeout is the timeout, in seconds, for database ping operations.
	PingTimeout int `mapstructure:"ping-timeout"`
	// MaxConns caps the connection pool size. 0 uses the pgx default (the
	// greater of 4 and runtime.NumCPU()).
	MaxConns int `mapstructure:"max-conns"`
	// MinConns is the number of idle connections kept warm in the pool. 0 uses
	// the pgx default (0).
	MinConns int `mapstructure:"min-conns"`
	// MaxConnLifetime is the duration, in seconds, after which a connection is
	// closed and replaced. 0 uses the pgx default (3600).
	MaxConnLifetime int `mapstructure:"max-conn-lifetime"`
	// MaxConnIdleTime is the duration, in seconds, after which an idle
	// connection is closed. 0 uses the pgx default (1800).
	MaxConnIdleTime int `mapstructure:"max-conn-idle-time"`
	// HealthCheckPeriod is the interval, in seconds, between pool health checks
	// of idle connections. 0 uses the pgx default (60).
	HealthCheckPeriod int `mapstructure:"health-check-period"`
	// CloseTimeout bounds, in seconds, how long shutdown waits for the pool to
	// close. pgxpool.Close blocks until every connection is returned and closed,
	// so a connection wedged mid-query would otherwise hang shutdown
	// indefinitely. 0 or less waits indefinitely (the raw pgxpool behaviour).
	CloseTimeout int `mapstructure:"close-timeout"`
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
