package hooks

import (
	"errors"
	"fmt"
)

type hook interface {
	NewHook(map[string]interface{}) (hook, error)
	Run(string, string) error
}

var hookTypes = map[string]hook{
	"download": downloadHook{},
	"git":      gitHook{},
}

func RunHooks(newVersion string, oldVersion string, hookConfigs []map[string]interface{}) error {
	hookRunners, err := getHooks(hookConfigs)
	if err != nil {
		return err
	}
	for _, hook := range hookRunners {
		err := hook.Run(newVersion, oldVersion)
		if err != nil {
			return err
		}
	}
	return nil
}

func getHooks(hookConfigs []map[string]interface{}) ([]hook, error) {
	var hooks []hook
	for _, config := range hookConfigs {
		if hookType, ok := config["type"]; ok {
			hookType := hookType.(string)
			if hook, ok := hookTypes[hookType]; ok {
				hook, err := hook.NewHook(config)
				if err != nil {
					return nil, err
				}
				hooks = append(hooks, hook)
			} else {
				return nil, errors.New(fmt.Sprintf("Hook '%s' not found", hookType))
			}
		} else {
			return nil, errors.New("Missing 'type' key in hook config")
		}
	}
	return hooks, nil
}
