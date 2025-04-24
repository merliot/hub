package device

import (
	"sort"
	"strings"
)

type serverFlags struct {
	flags
}

const (
	flagRunningDemo     flags = 1 << iota // Running in DEMO mode
	flagRunningSite                       // Running in SITE mode
	flagSaveToClipboard                   // Save changes to clipboard
	flagAutoSave                          // Automatically save changes
	flagDirty                             // Server has unsaved changes
	flagDebugKeepBuilds                   // Don't delete temp build directory
)

func (sf serverFlags) list() string {
	var list []string
	names := map[flags]string{
		flagRunningDemo:     "RUNNING_DEMO",
		flagRunningSite:     "RUNNING_SITE",
		flagSaveToClipboard: "SAVE_TO_CLIPBOARD",
		flagAutoSave:        "AUTO_SAVE",
		flagDirty:           "DIRTY",
		flagDebugKeepBuilds: "DEBUG_KEEP_BUILDS",
	}
	for flag, name := range names {
		if sf.isSet(flag) {
			list = append(list, name)
		}
	}
	sort.Strings(list)
	return strings.Join(list, "|")
}
