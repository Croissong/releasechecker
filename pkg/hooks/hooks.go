package hooks

import (
	"errors"
	"fmt"
	"github.com/croissong/releasechecker/pkg/log"
)

var hookMap = map[string]func(conf map[string]interface{}) (hook, error){
	"download": NewDownloader,
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
	for _, config := range hookConfigs {
		if hookType, ok := config["type"]; ok {
			hookType := hookType.(string)
			if hookConstructor, ok := hookMap[hookType]; ok {
				hook, err := hookConstructor(config)
				if err != nil {
					return nil, err
				}
				hooks = append(hooks, hook)
			}
			return nil, errors.New(fmt.Sprintf("Hook '%s' not found", hookType))
		}
		return nil, errors.New("Missing 'type' key in hook config")
	}
	return hooks, nil
}

type hook interface {
	Run(version string) error
}
