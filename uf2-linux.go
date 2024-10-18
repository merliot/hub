//go:build !tinygo

package hub

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/merliot/hub/target"
)

func GenerateUf2s(dir string) error {
	for name, model := range Models {
		var proto = &device{Model: name}
		proto.build(model.Maker)
		if err := proto.generateUf2s(dir); err != nil {
			return err
		}
	}
	return nil
}

func (d *device) generateUf2s(dir string) error {
	for _, target := range target.TinyGoTargets(d.Targets) {
		if err := d.generateUf2(dir, target); err != nil {
			return err
		}
	}
	return nil
}

func (d *device) generateUf2(dir, target string) error {

	// Create temp build directory
	temp, err := os.MkdirTemp("./", d.Model+"-")
	if err != nil {
		return err
	}

	if keepBuilds != "" {
		fmt.Println("DEBUG: Temporary build dir:", temp)
	} else {
		defer os.RemoveAll(temp)
	}

	var runnerGo = filepath.Join(temp, "runner.go")
	if err := d.genFile("device-runner-tinygo.tmpl", runnerGo, map[string]any{
		"model": Models[d.Model],
	}); err != nil {
		return err
	}

	// Build the uf2 file
	uf2Name := d.Model + "-" + target + ".uf2"
	output := filepath.Join(dir, uf2Name)
	cmd := exec.Command("tinygo", "build", "-target", target, "-o", output, "-stack-size", "8kb", "-size", "full", runnerGo)
	fmt.Println(cmd.String())
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}
	fmt.Println(string(stdoutStderr))

	return nil
}
