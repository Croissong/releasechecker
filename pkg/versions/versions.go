package versions

import (
	version "github.com/hashicorp/go-version"
	"sort"
)

func GetLatestVersion(versionStrings []string) (string, error) {
	var versions []*version.Version
	for _, vString := range versionStrings {
		v, err := version.NewVersion(vString)
		if err != nil {
			return "", err
		}
		versions = append(versions, v)
	}
	sort.Sort(version.Collection(versions))
	latestVersion := versions[len(versions)-1]
	return latestVersion.Original(), nil
}
