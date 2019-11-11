package hooks

import (
	"github.com/croissong/releasechecker/pkg/log"
)

var hookMap = map[string]hook{
	"download": downloader{},
}

func RunHooks(version string, hookConfigs []map[string]interface{}) error {
	hookRunners, err := getHooks(hookConfigs)
	if err != nil {
		return err
	}
	for _, hook := range hookRunners {
		err := hook.Run(version)
		if err != nil {
			log.Logger.Fatal(err)
			return err
		}
	}
	return nil
}

func getHooks(hookConfigs []map[string]interface{}) ([]hook, error) {
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
