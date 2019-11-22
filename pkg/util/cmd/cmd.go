package cmdutil

import (
	"fmt"
	"github.com/croissong/releasechecker/pkg/log"
	"io"
	"os/exec"
	"strings"
)

func RunCmd(command string, opts CmdOptions) (string, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("set -o pipefail; %s", command))

	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}

	log.Logger.Debug("Running cmd: ", cmd.String())

	if opts.Input != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Logger.Error(err)
			return "", err
		}

		go func() {
			defer stdin.Close()
			io.WriteString(stdin, opts.Input)
		}()
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Logger.Errorf("%s: %s", err, out)
		return "", err
	}

	output := stripWhitespace(string(out))
	return output, nil
}

type CmdOptions struct {
	Input string
	Dir   string
}

func stripWhitespace(str string) string {
	return strings.Join(strings.Fields(str), "")
}
