package providers

import (
	"errors"
	"fmt"
)

var providerMap = map[string]func(conf map[string]interface{}) (provider, error){
	"regex":  NewRegex,
	"github": NewGithub,
}

func GetProvider(providerConfig map[string]interface{}) (provider, error) {
	if providerType, ok := providerConfig["type"]; ok {
		providerType := providerType.(string)
		if provider, ok := providerMap[providerType]; ok {
			return provider(providerConfig)
		}
		return nil, errors.New(fmt.Sprintf("Provider '%s' not found", providerType))
	}
	return nil, errors.New("Missing 'type' key in provider config")
}

type provider interface {
	GetVersions() ([]string, error)
}
