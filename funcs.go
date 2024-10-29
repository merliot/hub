//go:build !tinygo

package hub

import (
	"encoding/json"
	"html/template"
	"slices"
	"strings"
	"time"

	"github.com/merliot/hub/target"
	"golang.org/x/exp/maps"
)

func linuxTarget(target string) bool {
	return target == "x86-64" || target == "rpi"
}

func (d *device) classOffline() string {
	if d.IsSet(flagOnline) {
		return ""
	} else {
		return "offline" // enables CSS class .offline
	}
}

func (d *device) stateJSON() (string, error) {
	bytes, err := json.MarshalIndent(d.State, "", "\t")
	return string(bytes), err
}

// funcs are device functions passed to templates.
//
// IMPORTANT!
//
// Don't add any functions that expose sensitive data such as passwd
//
// TODO how to split these into ones that all can use and ones that only core
// TODO template/ templates can use?
func (d *device) funcs() template.FuncMap {
	return template.FuncMap{
		"id":             func() string { return d.Id },
		"model":          func() string { return d.Model },
		"name":           func() string { return d.Name },
		"uniq":           func(s string) string { return d.Model + "-" + d.Id + "-" + s },
		"deployParams":   func() template.HTML { return d.DeployParams },
		"state":          func() any { return d.State },
		"stateJSON":      d.stateJSON,
		"uptime":         func() string { return time.Since(d.startup).Truncate(time.Second).String() },
		"title":          strings.Title,
		"add":            func(a, b int) int { return a + b },
		"mult":           func(a, b int) int { return a * b },
		"joinStrings":    func(parts ...string) string { return strings.Join(parts, "") },
		"contains":       func(s []string, v string) bool { return slices.Contains(s, v) },
		"targets":        func() target.Targets { return target.MakeTargets(d.Targets) },
		"ssids":          func() []string { return maps.Keys(wifiAuths()) },
		"target":         func() string { return d.deployValues().Get("target") },
		"port":           func() string { return d.deployValues().Get("port") },
		"ssid":           func() string { return d.deployValues().Get("ssid") },
		"package":        func() string { return Models[d.Model].Package },
		"source":         func() string { return Models[d.Model].Source },
		"isLinuxTarget":  linuxTarget,
		"isMissingWifi":  func() bool { return len(wifiAuths()) == 0 },
		"isRoot":         func() bool { return d == root },
		"isProgenitive":  func() bool { return d.IsSet(FlagProgenitive) },
		"wantsHttpPort":  func() bool { return d.IsSet(FlagWantsHttpPort) },
		"isOnline":       func() bool { return d.IsSet(flagOnline) },
		"isDemo":         func() bool { return d.IsSet(flagDemo) },
		"isDirty":        func() bool { return d.IsSet(flagDirty) },
		"isLocked":       func() bool { return d.IsSet(flagLocked) },
		"bgColor":        d.bgColor,
		"textColor":      d.textColor,
		"borderColor":    d.borderColor,
		"classOffline":   d.classOffline,
		"renderTemplate": d.renderTemplate,
		"renderView":     d.renderView,
		"renderChildren": d.renderChildren,
	}
}
