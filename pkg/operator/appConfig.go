package operator

type Config struct {
	KubeConfig            string
	DebugLevel            string
	WebServerPort         string
	DeleteSuccessPodDelay int
	DeleteFailedPodDelay  int
	KeepPods              bool
	Namespace             string
}

func GetDefaultConfig() Config {
	return Config{
		DebugLevel:            "INFO",
		KubeConfig:            "~/.kube/config",
		WebServerPort:         "8012",
		DeleteSuccessPodDelay: 60,
		DeleteFailedPodDelay:  300,
		KeepPods:              false,
		Namespace:             "default",
	}
}
