//go:build !tinygo

package device

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/merliot/hub/pkg/target"
	"github.com/merliot/hub/pkg/uf2"
)

func (s *server) generateUf2(d *device, dir, target string) error {

	// Create temp build directory
	temp, err := os.MkdirTemp("./", d.Model+"-")
	if err != nil {
		return err
	}

	if s.isSet(flagDebugKeepBuilds) {
		LogInfo("Temporary build", "dir", temp)
	} else {
		defer os.RemoveAll(temp)
	}

	var runnerFile = "runner.go"
	if err := d.genFile(temp, "device-runner-tinygo.tmpl", runnerFile, map[string]any{
		"model": d.model,
	}); err != nil {
		return err
	}

	// Build the uf2 file
	uf2Name := d.Model + "-" + target + ".uf2"
	output := filepath.Join(dir, uf2Name)
	input := filepath.Join(temp, runnerFile)
	cmd := exec.Command("tinygo", "build", "-target", target, "-o", output, "-stack-size", "8kb", "-size", "short", input)
	LogInfo(cmd.String())
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}
	LogInfo(string(stdoutStderr))

	return nil
}

func (s *server) Uf2GenerateBaseImages(dir, modelName, targetName string) (err error) {
	s.models.drange(func(name string, model *Model) bool {
		if name == modelName || modelName == "" {
			var proto = &device{
				Model: name,
				model: model,
			}
			proto.build(0)
			for _, target := range target.TinyGoTargets(proto.Targets) {
				if target == targetName || targetName == "" {
					if err = s.generateUf2(proto, dir, target); err != nil {
						return false
					}
				}
			}
		}
		return true
	})
	return
}

func Uf2Dump(file string) (string, error) {
	uf2, err := uf2.Read(file)
	if err != nil {
		return "", err
	}

	magicStart := fmt.Sprintf("{\"MagicStart\":\"%s\"", uf2Magic)
	magicEnd := fmt.Sprintf("\"MagicEnd\":\"%s\"}", uf2Magic)

	data := uf2.Bytes()
	start := bytes.Index(data, []byte(magicStart))
	if start == -1 {
		return "", fmt.Errorf("start uf2Magic sequence not found")
	}

	end := bytes.Index(data, []byte(magicEnd))
	if end == -1 {
		return "", fmt.Errorf("end uf2Magic sequence not found")
	}

	if end < start {
		return "", fmt.Errorf("uf2 search is messed up")
	}

	var block uf2ParamsBlock
	paramsBlock := data[start : end+len(magicEnd)]
	if err := json.Unmarshal(paramsBlock, &block); err != nil {
		return "", err
	}

	params, err := json.MarshalIndent(block.uf2Params, "", "\t")
	if err != nil {
		return "", err
	}

	return string(params), nil
}

func uf2Create(base, installer string, block uf2ParamsBlock) error {

	// Re-write the base uf2 file and save as the installer uf2 file.
	// The paramsMem area is replaced by json-encoded params.

	uf2, err := uf2.Read(base)
	if err != nil {
		return err
	}

	oldBytes := bytes.Repeat([]byte{byte('x')}, 2048)
	newBytes := make([]byte, 2048)

	newParams, err := json.Marshal(block)
	if err != nil {
		return err
	}
	copy(newBytes, newParams)

	uf2.ReplaceBytes(oldBytes, newBytes)

	if err = uf2.Write(installer); err != nil {
		return err
	}

	return nil
}

func Uf2Create(target, paramsJSON string) error {
	var params uf2Params
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return err
	}

	var p = uf2ParamsBlock{
		MagicStart: uf2Magic,
		uf2Params:  params,
		MagicEnd:   uf2Magic,
	}

	base := filepath.Join("bin", p.Model+"-"+target+".uf2")
	installer := p.Model + "-" + p.Id + "-installer.uf2"

	return uf2Create(base, installer, p)
}
