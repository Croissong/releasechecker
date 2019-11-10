package hooks

var hookMap = map[string]hook{
	"download": downloader{},
}

func GetHooks(hookConfigs []map[string]interface{}) ([]hook, error) {
	var hooks []hook
	for _, hookConfig := range hookConfigs {
		hookType := hookConfig["type"].(string)
		hook, err := hookMap[hookType].Init(hookConfig)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, hook)
	}
	return hooks, nil
}

type hook interface {
	Init(hookConfig map[string]interface{}) (hook, error)
	Run(version string) error
}
