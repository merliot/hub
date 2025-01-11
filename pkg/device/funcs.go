//go:build !tinygo

package device

import (
	"encoding/json"
	"html/template"
	"slices"
	"strings"
	"time"

	"github.com/merliot/hub/pkg/target"
	"golang.org/x/exp/maps"
)

func wantsWifi(target string) bool {
	return target == "pyportal" || target == "wioterminal" || target == "nano-rp2040"
}

func (d *device) classOffline() string {
	if d._isSet(flagOnline) {
		return ""
	} else {
		return "offline" // enables CSS class .offline
	}
}

func (d *device) stateJSON() (string, error) {
	bytes, err := json.MarshalIndent(d.State, "", "\t")
	return string(bytes), err
}

func aliveDevices() (alive deviceMap) {
	devicesMu.RLock()
	defer devicesMu.RUnlock()
	alive = make(deviceMap)
	for id, d := range devices {
		if !d.isSet(flagGhost) {
			alive[id] = d
		}
	}
	return
}

func devicesJSON() (string, error) {
	bytes, err := json.MarshalIndent(aliveDevices(), "", "\t")
	return string(bytes), err
}

func (d *device) uptime() string {
	return formatDuration(time.Since(d.startup).Truncate(time.Second))
}

// funcs are device functions passed to templates.
//
// IMPORTANT!
//
// Don't add any functions that expose sensitive data such as passwd
//
// TODO how to split these into ones that all can use and ones that only core
// TODO template/ templates can use?
func (d *device) baseFuncs() template.FuncMap {
	return template.FuncMap{
		"id":              func() string { return d.Id },
		"model":           func() string { return d.Model },
		"name":            func() string { return d.Name },
		"uniq":            func(s string) string { return d.Model + "-" + d.Id + "-" + s },
		"deployParams":    func() template.HTML { return d.DeployParams },
		"state":           func() any { return d.State },
		"stateJSON":       d.stateJSON,
		"uptime":          d.uptime,
		"title":           strings.Title,
		"add":             func(a, b int) int { return a + b },
		"mult":            func(a, b int) int { return a * b },
		"joinStrings":     func(parts ...string) string { return strings.Join(parts, "") },
		"contains":        func(s []string, v string) bool { return slices.Contains(s, v) },
		"targets":         func() target.Targets { return target.MakeTargets(d.Targets) },
		"ssids":           func() []string { return maps.Keys(wifiAuths()) },
		"target":          func() string { return d.deployValues().Get("target") },
		"port":            func() string { return d.deployValues().Get("port") },
		"ssid":            func() string { return d.deployValues().Get("ssid") },
		"package":         func() string { return Models[d.Model].Package },
		"devicesJSON":     devicesJSON,
		"isMissingWifi":   func() bool { return len(wifiAuths()) == 0 },
		"isRoot":          func() bool { return d == root },
		"isProgenitive":   func() bool { return d._isSet(FlagProgenitive) },
		"wantsHttpPort":   func() bool { return d._isSet(FlagWantsHttpPort) },
		"isOnline":        func() bool { return d._isSet(flagOnline) },
		"isDemo":          func() bool { return d._isSet(flagDemo) },
		"isDirty":         func() bool { return d._isSet(flagDirty) },
		"isLocked":        func() bool { return d._isSet(flagLocked) },
		"saveToClipboard": func() bool { return loadedFromDEVICES },
		"bgColor":         d.bgColor,
		"textColor":       d.textColor,
		"borderColor":     d.borderColor,
		"bodyColors":      bodyColors,
		"classOffline":    d.classOffline,
		"renderTemplate":  d.renderTemplate,
		"renderView":      d.renderView,
		"renderChildren":  d.renderChildren,
	}
}
