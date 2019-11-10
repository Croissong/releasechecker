package hooks

var hookMap = map[string]hook{
	"download": downloader{},
}

func GetHooks(hookConfigs []map[string]interface{}) []hook {
	var hooks []hook
	for _, hookConfig := range hookConfigs {
		hookType := hookConfig["type"].(string)
		hook := hookMap[hookType].Init(hookConfig)
		hooks = append(hooks, hook)
	}
	return hooks
}

type hook interface {
	Init(hookConfig map[string]interface{}) hook
	Run(version string) error
}
