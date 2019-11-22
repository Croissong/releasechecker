package docker

import (
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestDocker(t *testing.T) {
	config := map[string]interface{}{
		"repo": "prom/prometheus",
	}
	docker, err := Docker{}.NewProvider(config)
	assert.Equal(t, err, nil)
	versions, err := docker.GetVersions()
	assert.Equal(t, err, nil)
	assert.Equal(t, len(versions) > 0, true)
}
