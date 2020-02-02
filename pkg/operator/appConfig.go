package operator

type Config struct {
	KubeConfig string
	DebugLevel string
}

func GetDefaultConfig() Config{
	return Config{
		DebugLevel: "INFO",
		KubeConfig: "~/.kube/config",
	}
}