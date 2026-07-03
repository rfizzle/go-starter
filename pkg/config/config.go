package config

type Application struct {
	Name      string    `json:"name"`
	Build     Build     `json:"build"`
	Webserver Webserver `json:"webserver"`
	Database  Database  `json:"database"`
	Auth      Auth      `json:"auth"`
	// ShutdownDelay is the delay, in seconds, between receiving a shutdown
	// signal and beginning graceful shutdown. It gives orchestrators such as
	// Kubernetes time to propagate the not-ready state and stop routing traffic
	// before in-flight components are torn down. Default 0 (no delay).
	ShutdownDelay int `json:"shutdown-delay"`
}

type Build struct {
	Branch   string `json:"branch"`
	Date     string `json:"date"`
	Env      string `json:"env"`
	Revision string `json:"revision"`
	Version  string `json:"version"`
}

type Webserver struct {
	Address      string `json:"address"`
	Port         int    `json:"port"`
	ReadTimeout  int    `json:"read-timeout"`
	WriteTimeout int    `json:"write-timeout"`
	IdleTimeout  int    `json:"idle-timeout"`
}

type Database struct {
	// URI is the PostgreSQL connection string, e.g.
	// "postgres://user:password@localhost:5432/dbname".
	URI string `json:"uri"`
	// PingTimeout is the timeout, in seconds, for database ping operations.
	PingTimeout int `json:"ping-timeout"`
}

type Auth struct {
	// JWTSecret is the shared HMAC (HS256) secret used to sign and verify JWT
	// bearer tokens. When empty, protected endpoints fail closed (unauthorized).
	JWTSecret string `json:"jwt-secret"`
	// JWTIssuer, when set, is required on validated tokens ("iss" claim).
	JWTIssuer string `json:"jwt-issuer"`
	// JWTAudience, when set, is required on validated tokens ("aud" claim).
	JWTAudience string `json:"jwt-audience"`
	// JWTTTL is the lifetime, in seconds, of generated tokens. Default 3600.
	JWTTTL int `json:"jwt-ttl"`
}
