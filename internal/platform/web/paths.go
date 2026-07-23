package web

func ignorePathsLogging() []string {
	return []string{
		"/metrics",
		"/debug/*",
		"/api/v1/healthz/ready",
		"/api/v1/healthz/live",
		"/favicon.ico",
	}
}
