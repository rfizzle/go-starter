package config

type Application struct {
	Name      string    `json:"name"`
	Build     Build     `json:"build"`
	Webserver Webserver `json:"webserver"`
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
