package providers

import (
	"errors"
	"fmt"
)

var providers = map[string]func(conf map[string]interface{}) (provider, error){
	"command": NewCommand,
	"github":  NewGithub,
	"regex":   NewRegex,
}

func GetProvider(providerConfig map[string]interface{}) (provider, error) {
	if providerType, ok := providerConfig["type"]; ok {
		providerType := providerType.(string)
		if provider, ok := providers[providerType]; ok {
			return provider(providerConfig)
		}
		return nil, errors.New(fmt.Sprintf("Provider '%s' not found", providerType))
	}
	return nil, errors.New("Missing 'type' key in provider config")
}

type provider interface {
	GetVersion() (string, error)
	GetVersions() ([]string, error)
}
