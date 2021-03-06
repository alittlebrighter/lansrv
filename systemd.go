/*
Package lansrv: These functions are built to work with systemd.
*/
package lansrv

import (
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
func ParseServiceFiles(configFiles []string) []LanAd {
	configs := []LanAd{}

	for _, configFile := range configFiles {
		ini := goini.New()
		if err := ini.ParseFile(configFile); err != nil {
			continue
		}

		adMap, ok := ini.GetKvmap("LanSrv")
		if !ok {
			continue
		}

		ad := new(LanAd)
		if err := ad.FromMap(adMap); err == nil {
			fmt.Println("adding lan ad:", adMap)
			configs = append(configs, *ad)
		}
	}

	return configs
}
