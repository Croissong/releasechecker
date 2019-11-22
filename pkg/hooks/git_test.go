package hooks

import "testing"

func TestClone(t *testing.T) {
	config := map[string]interface{}{
		"repo": "git@github.com:Croissong/verdun.git",
		"change": map[string]interface{}{
			"command": "yq w k8s/values/images.yml 'prometheus.tag' {{.NewVersion}} | sponge k8s/values/images.yml",
		},
		"commit": map[string]interface{}{
			"msgTemplate": "Bump prometheus -> {{.NewVersion}}",
			"branch":      "master",
			"push":        false,
			"authorEmail": "releasewatcher@patrician.cloud",
			"authorName":  "ReleaseWatcher Bot",
		},
	}
	git, _ := NewGitHook(config)
	git.Run("v2.14.0", "v2.13.1")
}
