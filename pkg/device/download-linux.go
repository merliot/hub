//go:build !tinygo

package device

import (
	"net/http"
	"path/filepath"
	"strings"

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
	referer := r.Referer()
	service := d.Model + "-" + d.Id
	dialurls := strings.Replace(referer, "http", "ws", 1) + "ws"

	// Generate environment variable file.  The service will load env vars
	// from this file.
	if err := d.genFile(dir, "device-env.tmpl", "env", map[string]any{
		"port":            r.URL.Query().Get("port"),
		"user":            s.user,
		"passwd":          s.passwd,
		"dialurls":        dialurls,
		"logLevel":        s.logLevel,
		"pingPeriod":      s.wsxPingPeriod,
		"background":      s.background,
		"autoSave":        s.isSet(flagAutoSave),
		"wifiSsids":       strings.Join(s.wifiSsids, ","),
		"wifiPassphrases": strings.Join(s.wifiPassphrases, ","),
	}); err != nil {
		return err
	}

	// Generate systemd merliot.target unit from
	// device-merliot-target.tmpl.  This will be the parent unit of all
	// device units.
	targetFile := "merliot.target"
	if err := d.genFile(dir, "device-merliot-target.tmpl", targetFile, nil); err != nil {
		return err
	}

	// Generate systemd {{service}}.service unit from device-service.tmpl
	serviceFile := service + ".service"
	if err := d.genFile(dir, "device-service.tmpl", serviceFile, map[string]any{
		"service": service,
	}); err != nil {
		return err
	}

	// Generate {{service}}.conf from device-conf.tmpl.  This sets up
	// logging service for the device.  Logs are available at
	// /var/log/{{.service}}.log.
	confFile := service + ".conf"
	if err := d.genFile(dir, "device-conf.tmpl", confFile, map[string]any{
		"service": service,
	}); err != nil {
		return err
	}

	// Generate SelF-eXtracting (SFX) installer script
	sfxFile := "sfx.sh"
	if err := d.genFile(dir, "device-sfx.tmpl", sfxFile, map[string]any{
		"service": service,
	}); err != nil {
		return err
	}

	// Generate service install script
	if err := d.genFile(dir, "device-install.tmpl", "install.sh", map[string]any{
		"service": service,
	}); err != nil {
		return err
	}

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
	if err := createSFX(dir, sfxFile, tarFile, installer); err != nil {
		return err
	}

	// Serve installer file for download
	return s.serveFile(w, r, filepath.Join(dir, installer))
}
