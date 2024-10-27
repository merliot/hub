//go:build !tinygo

package hub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/merliot/hub/target"
	"github.com/merliot/hub/uf2"
)

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
		slog.Debug("Temporary build", "dir", temp)
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
	slog.Debug(cmd.String())
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stdoutStderr)
	}
	slog.Debug(string(stdoutStderr))

	return nil
}

func Uf2GenerateBaseImages(dir string) error {
	for name, model := range Models {
		var proto = &device{Model: name}
		proto.build(model.Maker)
		if err := proto.generateUf2s(dir); err != nil {
			return err
		}
	}
	return nil
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

	base := filepath.Join("uf2s", p.Model+"-"+target+".uf2")
	installer := p.Model + "-" + p.Id + "-installer.uf2"

	return uf2Create(base, installer, p)
}
