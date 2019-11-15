package sources

import (
	"errors"
	"fmt"
)

var sourceMap = map[string]func(conf map[string]interface{}) (source, error){
	"command": NewCommand,
}

func GetSource(sourceConfig map[string]interface{}) (source, error) {
	if sourceType, ok := sourceConfig["type"]; ok {
		sourceType := sourceType.(string)
		if source, ok := sourceMap[sourceType]; ok {
			return source(sourceConfig)
		}
		return nil, errors.New(fmt.Sprintf("Source '%s' not found", sourceType))
	}
	return nil, errors.New("Missing 'type' key in source config")
}

type source interface {
	GetVersion() (string, error)
}
