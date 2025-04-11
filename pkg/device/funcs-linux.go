//go:build !tinygo

package device

import (
	"encoding/json"
	"html/template"
	"slices"
	"strings"
	"time"

	"github.com/merliot/hub/pkg/target"
)

func tinygoTarget(target string) bool {
	return target == "pyportal" || target == "wioterminal" || target == "nano-rp2040"
}

func (s *server) bodyColors() string {
	if s.background == "GOOD" {
		return "bg-space-white text"
	}
	return "bg-black text"
}

// Base server funcs are functions passed to all templates.
//
// IMPORTANT!
//
// Don't add any functions that expose sensitive data such as passwd
func (s *server) baseFuncs() FuncMap {
	return FuncMap{
		"saveToClipboard": func() bool { return s.isSet(flagSaveToClipboard) },
		"devicesJSON":     func() string { return string(s.devices.getPrettyJSON()) },
		"title":           strings.Title,
		"add":             func(a, b int) int { return a + b },
		"mult":            func(a, b int) int { return a * b },
		"joinStrings":     func(parts ...string) string { return strings.Join(parts, "") },
		"contains":        func(s []string, v string) bool { return slices.Contains(s, v) },
		"tinygoTarget":    tinygoTarget,
		"ssids":           func() []string { return s.wifiSsids },
		"isMissingWifi":   func() bool { return len(s.wifiSsids) == 0 },
		"bodyColors":      s.bodyColors,
		"isDirty":         func() bool { return s.isSet(flagDirty) },
	}
}

func (d *device) classOffline() string {
	if d.isSet(flagOnline) {
		return ""
	} else {
		return "offline" // enables CSS class .offline
	}
}

func (d *device) stateJSON() []byte {
	bytes, _ := json.MarshalIndent(d.State, "", "\t")
	return bytes
}

func (d *device) uptime() string {
	return formatDuration(time.Since(d.startup).Truncate(time.Second))
}

// Base device funcs are functions passed to device templates.
//
// IMPORTANT!
//
// Don't add any functions that expose sensitive data such as passwd
func (d *device) baseFuncs() FuncMap {
	return FuncMap{
		"id":             func() string { return d.Id },
		"model":          func() string { return d.Model },
		"name":           func() string { return d.Name },
		"uniq":           func(s string) string { return d.Model + "-" + d.Id + "-" + s },
		"deployParams":   func() template.URL { return template.URL(d.DeployParams) },
		"state":          func() any { return d.State },
		"stateJSON":      func() string { return string(d.stateJSON()) },
		"uptime":         d.uptime,
		"targets":        func() target.Targets { return target.MakeTargets(d.Targets) },
		"target":         func() string { return d.deployValues().Get("target") },
		"port":           func() string { return d.deployValues().Get("port") },
		"ssid":           func() string { return d.deployValues().Get("ssid") },
		"package":        func() string { return d.model.Package },
		"isRoot":         func() bool { return d.isSet(flagRoot) },
		"isProgenitive":  func() bool { return d.isSet(FlagProgenitive) },
		"isHttpPortMust": func() bool { return d.isSet(FlagHttpPortMust) },
		"isOnline":       func() bool { return d.isSet(flagOnline) },
		"isLocked":       func() bool { return d.isSet(flagLocked) },
		"bgColor":        d.bgColor,
		"textColor":      d.textColor,
		"borderColor":    d.borderColor,
		"classOffline":   d.classOffline,
		"renderTemplate": d.renderTemplate,
		"renderView":     d.renderView,
		"renderChildren": d.renderChildren,
	}
}
