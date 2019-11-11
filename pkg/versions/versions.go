package versions

import (
	version "github.com/hashicorp/go-version"
	"sort"
)

func IsNewer(a *version.Version, b *version.Version) bool {
	return a.Compare(b) == 1
}

func GetLatestVersion(versionStrings []string) (*version.Version, error) {
	var versions []*version.Version
	for _, vString := range versionStrings {
		v, err := version.NewVersion(vString)
		if err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	sort.Sort(version.Collection(versions))
	latestVersion := versions[len(versions)-1]
	return latestVersion, nil
}
