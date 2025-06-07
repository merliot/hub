//go:build !tinygo

package device

import (
	"net/http"
	"path/filepath"

	tpkg "github.com/merliot/hub/pkg/target"
)

// collectBinFiles determines which binaries to include in the installer
func (s *server) collectBinFiles(d *device, target string) []string {
	var binFiles []string
	childModels := s.childModels(d)
	if childModels.length() == 0 {
		// Sterile device only needs the bin/device-<target> binary
		binFiles = append(binFiles, "-C", ".", "./bin/device-"+target)
	} else {
		// Copy over the binaries needed to produce children devices
		binFiles = append(binFiles, "-C", ".", "./bin/device-rpi")
		binFiles = append(binFiles, "-C", ".", "./bin/device-x86-64")
		childModels.drange(func(name string, model *Model) bool {
			// Copy the UF2 files for the model (all targets)
			for _, t := range tpkg.TinyGoTargets(model.Config.Targets) {
				binFiles = append(binFiles, "-C", ".", "./bin/"+name+"-"+t+".uf2")
			}
			return true
		})
	}
	return binFiles
}

func (s *server) buildLinuxImage(d *device, w http.ResponseWriter, r *http.Request, dir, target string) error {
	// referer := r.Referer() // No longer used
	service := d.Model + "-" + d.Id
	// dialurls := strings.Replace(referer, "http", "ws", 1) + "ws" // No longer used

	// TODO: Reimplement file generation for env, target, service, conf, sfx, and install scripts using Go or gomponents.
	// The following template-based file generation has been removed:
	// if err := d.genFile(dir, "device-env.tmpl", "env", ...); err != nil { return err }
	// if err := d.genFile(dir, "device-merliot-target.tmpl", ...); err != nil { return err }
	// if err := d.genFile(dir, "device-service.tmpl", ...); err != nil { return err }
	// if err := d.genFile(dir, "device-conf.tmpl", ...); err != nil { return err }
	// if err := d.genFile(dir, "device-sfx.tmpl", ...); err != nil { return err }
	// if err := d.genFile(dir, "device-install.tmpl", ...); err != nil { return err }

	// Make a devices.json file
	if err := fileWriteJSON(filepath.Join(dir, "devices.json"), d.familyTree()); err != nil {
		return err
	}

	// Create a gzipped tar ball with everything inside need to
	// install/uninstall the device
	tarFile := service + ".tar.gz"
	if err := s.createTarBall(dir, tarFile, s.collectBinFiles(d, target)); err != nil {
		return err
	}

	// Create final SFX image as the installer
	installer := service + "-installer"
	sfxFile := "sfx.sh"
	if err := createSFX(dir, sfxFile, tarFile, installer); err != nil {
		return err
	}

	// Serve installer file for download
	return s.serveFile(w, r, filepath.Join(dir, installer))
}
