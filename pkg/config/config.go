package config

type Application struct {
	Name      string    `json:"name"`
	Build     Build     `json:"build"`
	Webserver Webserver `json:"webserver"`
	Database  Database  `json:"database"`
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
