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
	RosterNames         map[int32]string
	RankShorts          map[int32]string
	RecordTypeNames     map[int32]string
	UniformTriggerTypes []int
	AOR                 map[string]string
	AORSubstrings       map[string]string
	AwolActive          []int
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

	if v := viper.GetString("RANK_SHORT_NAMES"); v != "" {
		var ranks map[int32]string
		if err := json.Unmarshal([]byte(v), &ranks); err == nil {
			for k, val := range ranks {
				om.RankShorts[k] = val
			}
		}
	}

	if v := viper.GetString("RECORD_TYPE_NAMES"); v != "" {
		var records map[int32]string
		if err := json.Unmarshal([]byte(v), &records); err == nil {
			for k, val := range records {
				om.RecordTypeNames[k] = val
			}
		}
	}

	if v := viper.GetString("UNIFORM_TRIGGER_RECORD_TYPES"); v != "" {
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
			om.UniformTriggerTypes = ids
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
			1: "ROSTER_TYPE_ACTIVE",
			2: "ROSTER_TYPE_RESERVE",
			3: "ROSTER_TYPE_EXTENDED_LEAVE",
			4: "ROSTER_TYPE_HONOR_ROLL",
			5: "ROSTER_TYPE_MEMORIAL",
			6: "ROSTER_TYPE_FORMER_MEMBERS",
		},
		RankShorts: map[int32]string{
			0:  "UNSPECIFIED",
			1:  "GOA",
			2:  "GEN",
			3:  "LTG",
			4:  "MG",
			5:  "BG",
			6:  "COL",
			7:  "LTC",
			8:  "MAJ",
			9:  "CPT",
			10: "1LT",
			11: "2LT",
			12: "CSM",
			13: "SGM",
			14: "1SG",
			15: "MSG",
			16: "SFC",
			17: "SSG",
			18: "SGT",
			19: "CPL",
			20: "SPC",
			21: "PFC",
			22: "PVT",
			23: "RCT",
			26: "CW5",
			27: "CW4",
			28: "CW3",
			29: "CW2",
			30: "WO1",
			31: "AR",
		},
		RecordTypeNames: map[int32]string{
			0: "RECORD_TYPE_UNSPECIFIED",
			1: "RECORD_TYPE_PROMOTION",
			2: "RECORD_TYPE_OPERATION",
			3: "RECORD_TYPE_TRANSFER",
			4: "RECORD_TYPE_DISCIPLINARY",
			5: "RECORD_TYPE_DISCHARGE",
			6: "RECORD_TYPE_ASSIGNMENT",
			7: "RECORD_TYPE_NAME_CHANGE",
			8: "RECORD_TYPE_ELOA",
			9: "RECORD_TYPE_GRADUATION",
		},
		UniformTriggerTypes: []int{1, 6, 7, 8, 9},
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

// RankShort returns the configured short name for a rank id. Uncataloged ids
// fall back to their decimal string so the wire stays deterministic.
func RankShort(rankId int32) string {
	ensureLoaded()
	orgMu.RLock()
	defer orgMu.RUnlock()
	if short, ok := org.RankShorts[rankId]; ok {
		return short
	}
	return strconv.FormatInt(int64(rankId), 10)
}

// RecordTypeName returns the wire name for a record type id. Uncataloged ids
// fall back to their decimal string.
func RecordTypeName(recordTypeId int32) string {
	ensureLoaded()
	orgMu.RLock()
	defer orgMu.RUnlock()
	if name, ok := org.RecordTypeNames[recordTypeId]; ok {
		return name
	}
	return strconv.FormatInt(int64(recordTypeId), 10)
}

// RecordTypeIsCataloged reports whether a record type id has a wire name.
func RecordTypeIsCataloged(recordTypeId int32) bool {
	ensureLoaded()
	orgMu.RLock()
	defer orgMu.RUnlock()
	_, ok := org.RecordTypeNames[recordTypeId]
	return ok
}

// IsUniformTriggerRecordType reports whether a record type should trigger the
// S1 uniform update recalculation.
func IsUniformTriggerRecordType(recordTypeId int32) bool {
	ensureLoaded()
	orgMu.RLock()
	defer orgMu.RUnlock()
	for _, id := range org.UniformTriggerTypes {
		if int32(id) == recordTypeId {
			return true
		}
	}
	return false
}
