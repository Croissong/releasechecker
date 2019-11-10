package providers

var providerMap = map[string]provider{
	"regex": regex{},
}

func GetProvider(providerConfig map[string]interface{}) provider {
	providerType := providerConfig["type"].(string)
	provider := providerMap[providerType].Init(providerConfig)
	return provider
}

type provider interface {
	Init(providerConfig map[string]interface{}) provider
	GetVersions() []string
}
