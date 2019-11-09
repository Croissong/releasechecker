package hooks

import (
	"bytes"
	"fmt"
	"github.com/croissong/releasechecker/pkg/log"
	"github.com/mholt/archiver"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"
)

func DownloadFile(version string, urlTemplate string, targetPath string, extract bool) error {
	url, err := buildUrl(version, urlTemplate)
	urlParts := strings.Split(url, "/")
	downloadPath := urlParts[len(urlParts)-1]
	if err != nil {
		return err
	}

	log.Logger.Info("Downloading ", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	tmpfile, err := ioutil.TempFile("", fmt.Sprintf("*%s", downloadPath))
	log.Logger.Debug("Downloading to ", tmpfile.Name())
	if err != nil {
		log.Logger.Fatal(err)
	}

	defer os.Remove(tmpfile.Name()) // clean up

	if _, err = io.Copy(tmpfile, resp.Body); err != nil {
		log.Logger.Fatal(err)
		return err
	}
	if err := tmpfile.Close(); err != nil {
		log.Logger.Fatal(err)
		return err
	}

	if extract {
		log.Logger.Info("Extracting archive ", tmpfile.Name())
		err = archiver.Unarchive(tmpfile.Name(), targetPath)
		if err != nil {
			log.Logger.Fatal(err)
		}
	} else {
		os.Rename(tmpfile.Name(), targetPath)
	}
	return nil
}

func buildUrl(version string, urlTemplate string) (string, error) {
	tmpl, err := template.New("urlTemplate").Parse(urlTemplate)
	if err != nil {
		return "", err
	}
	data := struct {
		Version string
	}{
		Version: version,
	}
	var tpl bytes.Buffer
	tmpl.Execute(&tpl, data)

	return tpl.String(), nil
}
