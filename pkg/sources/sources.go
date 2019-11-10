package sources

import (
	"fmt"
	"github.com/croissong/releasechecker/pkg/log"
	"os/exec"
)

var sourceMap = map[string]source{
	"command": command{},
}

func GetSource(sourceConfig map[string]interface{}) source {
	sourceType := sourceConfig["type"].(string)
	source := sourceMap[sourceType].Init(sourceConfig)
	return source
}

type source interface {
	Init(sourceConfig map[string]interface{}) source
	GetVersion() string
}

type command struct {
	command string
}

func (cmd command) Init(sourceConfig map[string]interface{}) source {
	cmd.command = sourceConfig["command"].(string)
	return cmd
}

func (cmd command) GetVersion() string {
	sourceCmd := exec.Command("bash", "-c", fmt.Sprintf("%s", cmd.command))
	log.Logger.Debug("Running source cmd: ", sourceCmd.String())
	out, err := sourceCmd.Output()
	if err != nil {
		log.Logger.Error(err)
	}
	return string(out)
}
