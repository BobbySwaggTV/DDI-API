// Package config centralizes DDI organization-specific mappings that were
// previously hardcoded in the API. Defaults are the current 7CAV values so
// existing test fixtures and deployments keep working without env changes.
package config

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// orgMap holds the runtime organization mapping state.
type orgMap struct {
	RosterNames   map[int32]string
	AOR           map[string]string
	AORSubstrings map[string]string
	AwolActive    []int
}

var (
	orgMu sync.RWMutex
	org   *orgMap
)

// Load reads the organization mapping env vars and caches the result.
// Call this once after viper has been initialized (cmd/root.go does this).
// If a key is missing, the 7CAV default is used.
func Load() {
	orgMu.Lock()
	defer orgMu.Unlock()
	org = loadFromViper()
}

// ensureLoaded lazily loads defaults on first access. This lets tests and
// code paths that run before cmd/root.go still function.
func ensureLoaded() {
	orgMu.RLock()
	if org != nil {
		orgMu.RUnlock()
		return
	}
	orgMu.RUnlock()

	orgMu.Lock()
	defer orgMu.Unlock()
	if org != nil {
		return
	}
	org = defaultOrgMap()
}

func loadFromViper() *orgMap {
	om := defaultOrgMap()

	if v := viper.GetString("ORG_ROSTER_NAMES"); v != "" {
		var names map[int32]string
		if err := json.Unmarshal([]byte(v), &names); err == nil {
			for k, val := range names {
				om.RosterNames[k] = val
			}
		}
	}

	if v := viper.GetString("ORG_POSITION_GROUP_AOR"); v != "" {
		var aor map[string]string
		if err := json.Unmarshal([]byte(v), &aor); err == nil {
			for k, val := range aor {
				om.AOR[k] = val
			}
		}
	}

	if v := viper.GetString("ORG_AOR_SUBSTRINGS"); v != "" {
		var subs map[string]string
		if err := json.Unmarshal([]byte(v), &subs); err == nil {
			for k, val := range subs {
				om.AORSubstrings[k] = val
			}
		}
	}

	if v := viper.GetString("ORG_AWOL_ACTIVE_ROSTERS"); v != "" {
		var ids []int
		for _, s := range strings.Split(v, ",") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			if i, err := strconv.Atoi(s); err == nil {
				ids = append(ids, i)
			}
		}
		if len(ids) > 0 {
			om.AwolActive = ids
		}
	}

	return om
}

func defaultOrgMap() *orgMap {
	return &orgMap{
		RosterNames: map[int32]string{
			0: "ROSTER_TYPE_UNSPECIFIED",
			1: "ROSTER_TYPE_COMBAT",
			2: "ROSTER_TYPE_RESERVE",
			3: "ROSTER_TYPE_ELOA",
			4: "ROSTER_TYPE_WALL_OF_HONOR",
			5: "ROSTER_TYPE_ARLINGTON",
			6: "ROSTER_TYPE_PAST_MEMBERS",
		},
		AOR: map[string]string{
			"Regimental HQ":             "HHQ",
			"Support Attachment":        "HHQ",
			"New Recruits":              "RTC",
			"7th Cavalry Reservists":    "Reserve",
			"Extended Leave Of Absence": "ELOA",
		},
		AORSubstrings: map[string]string{
			"Command": "HHQ",
		},
		AwolActive: []int{1, 2},
	}
}

// RosterName returns the wire name for a roster id. Uncataloged ids fall back
// to their decimal string.
func RosterName(rt int32) string {
	ensureLoaded()
	orgMu.RLock()
	defer orgMu.RUnlock()
	if n, ok := org.RosterNames[rt]; ok {
		return n
	}
	return strconv.FormatInt(int64(rt), 10)
}

// RosterValue returns the roster id for a wire name.
func RosterValue(name string) (int32, bool) {
	ensureLoaded()
	orgMu.RLock()
	defer orgMu.RUnlock()
	for k, v := range org.RosterNames {
		if v == name {
			return k, true
		}
	}
	return 0, false
}

// RosterIsCataloged reports whether a numeric roster id is defined.
func RosterIsCataloged(rt int32) bool {
	ensureLoaded()
	orgMu.RLock()
	defer orgMu.RUnlock()
	_, ok := org.RosterNames[rt]
	return ok
}

// CatalogedRosterIDs returns the sorted list of defined roster ids.
func CatalogedRosterIDs() []int32 {
	ensureLoaded()
	orgMu.RLock()
	defer orgMu.RUnlock()
	var ids []int32
	for k := range org.RosterNames {
		ids = append(ids, k)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// AORFor returns the area-of-responsibility for a position group title. It
// first checks the exact map, then the substring map, and finally returns the
// original title as its own AOR.
func AORFor(title string) string {
	ensureLoaded()
	orgMu.RLock()
	defer orgMu.RUnlock()

	if aor, ok := org.AOR[title]; ok {
		return aor
	}

	for substr, aor := range org.AORSubstrings {
		if strings.Contains(title, substr) {
			return aor
		}
	}

	return title
}

// AwolActiveRosters returns the roster ids that should be considered active for
// the AWOL report.
func AwolActiveRosters() []int {
	ensureLoaded()
	orgMu.RLock()
	defer orgMu.RUnlock()
	return org.AwolActive
}
