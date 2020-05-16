/*
Package lansrv: These functions are built to work with systemd.
*/
package lansrv

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zieckey/goini"
)

// GatherServiceConfigs accepts @arg dir which it will walk recusively and collect an array
// of file names that contain a lansrv config.  Lansrv is built to work with systemd so any file
// ending with `.service` will be included in the array.
func GatherServiceConfigs(dir string) (configFiles []string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, _ error) error {
		if strings.HasSuffix(info.Name(), ".service") {
			configFiles = append(configFiles, path)
		}

		return nil
	})

	return
}

// ParseServiceConfigs takes @arg configs (list of filenames), tries to parse them
// as ini files and returns all non-nil results for lansrv configurations containing at least a Name
// and a Port.
func ParseServiceFiles(configFiles []string) (configs []LanAd) {
	for _, configFile := range configFiles {
		ini := goini.New()
		if err := ini.ParseFile(configFile); err != nil {
			continue
		}

		adMap, ok := ini.GetKvmap(service)
		if !ok {
			continue
		}

		ad, err := NewLanAd(adMap)
		if err != nil {
			continue
		}

		configs = append(configs, ad)
	}

	return
}

type LanAd map[string]string

func NewLanAd(kvMap map[string]string) (LanAd, error) {
	name, nameOk := kvMap["Name"]
	port, portOk := kvMap["Port"]

	if !nameOk || !portOk {
		return nil, errors.New(fmt.Sprint("Name or Port not found: Name=", name, " Port=", port))
	}

	return LanAd(kvMap), nil
}

func (la LanAd) Name() string {
	name, _ := la["Name"]
	return name
}

func (la LanAd) Port() string {
	port, _ := la["Port"]
	return port
}
