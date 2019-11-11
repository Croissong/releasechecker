package sources

import (
	"bytes"
	"fmt"
	"github.com/croissong/releasechecker/pkg/log"
	"github.com/croissong/releasechecker/pkg/util"
	ver "github.com/hashicorp/go-version"
	"os/exec"
	"strings"
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
	GetVersion() (*ver.Version, error)
}

type command struct {
	command string
}

func (cmd command) Init(sourceConfig map[string]interface{}) source {
	cmd.command = sourceConfig["command"].(string)
	return cmd
}

func (cmd command) GetVersion() (*ver.Version, error) {
	sourceCmd := exec.Command("bash", "-c", fmt.Sprintf("set -o pipefail; %s", cmd.command))
	log.Logger.Debug("Running source cmd: ", sourceCmd.String())
	var out bytes.Buffer
	sourceCmd.Stdout = &out
	sourceCmd.Stderr = &out
	err := sourceCmd.Run()
	if err != nil {
		if strings.Contains(out.String(), "command not found") {
			return nil, nil
		}
		return nil, err
	}
	log.Logger.Debugf("Running source cmd: %s", out.String())
	versionString := util.StripWhitespace(out.String())
	version, err := ver.NewVersion(versionString)
	if err != nil {
		return nil, err
	}
	return version, nil
}
