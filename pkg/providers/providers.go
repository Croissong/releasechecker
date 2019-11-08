package providers

import (
	"github.com/croissong/releasechecker/pkg/log"
	"io/ioutil"
	"net/http"
)

func GetVersion() {
	resp, err := http.Get("https://releases.hashicorp.com/terraform/")
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	log.Logger.Infof("%s", body)
}
