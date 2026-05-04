package oauth

// KuveraCredentials holds Kuvera platform login credentials.
type KuveraCredentials struct {
	Username string
	Password string
}

// StockalCredentials holds Stockal platform login credentials.
type StockalCredentials struct {
	Username string
	Password string
}

// Credentials holds credentials for all supported platforms.
type Credentials struct {
	Kuvera  *KuveraCredentials
	Stockal *StockalCredentials
}
