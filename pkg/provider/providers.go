package provider

import (
	"errors"
	"fmt"
	"github.com/hashicorp/go-version"
	"sort"
)

type Provider interface {
	NewProvider(map[string]interface{}) (Provider, error)
	GetVersion() (*version.Version, error)
	GetVersions() ([]*version.Version, error)
}

func GetProvider(providers map[string]Provider, providerConfig map[string]interface{}) (Provider, error) {
	if providerType, ok := providerConfig["type"]; ok {
		providerType := providerType.(string)
		if provider, ok := providers[providerType]; ok {
			return provider.NewProvider(providerConfig)
		}
		return nil, errors.New(fmt.Sprintf("Provider '%s' not found", providerType))
	}
	return nil, errors.New("Missing 'type' key in provider config")
}

func GetLatestVersion(provider Provider) (*version.Version, error) {
	versions, err := provider.GetVersions()
	if err != nil {
		return nil, err
	}
	sort.Sort(version.Collection(versions))
	latestVersion := versions[len(versions)-1]
	return latestVersion, nil
}

func IsNewerVersion(a *version.Version, b *version.Version) bool {
	return a.Compare(b) == 1
}
