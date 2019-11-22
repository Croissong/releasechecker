package yaml

import (
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestYaml(t *testing.T) {
	config := map[string]interface{}{
		"url":  "https://raw.githubusercontent.com/Croissong/verdun/master/k8s/values/images.yml",
		"path": "prometheus.tag",
	}
	yaml, _ := Yaml{}.NewProvider(config)
	version, _ := yaml.GetVersion()
	assert.Equal(t, version, "v2.13.1")
}
