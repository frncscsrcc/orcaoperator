package operator

type Config struct {
	KubeConfig    string
	DebugLevel    string
	WebServerPort string
}

func GetDefaultConfig() Config {
	return Config{
		DebugLevel:    "INFO",
		KubeConfig:    "~/.kube/config",
		WebServerPort: "8012",
	}
}
