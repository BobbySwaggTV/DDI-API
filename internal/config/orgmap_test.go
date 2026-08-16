package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRosterNameDefaults(t *testing.T) {
	Load()
	assert.Equal(t, "ROSTER_TYPE_UNSPECIFIED", RosterName(0))
	assert.Equal(t, "ROSTER_TYPE_ACTIVE", RosterName(1))
	assert.Equal(t, "ROSTER_TYPE_RESERVE", RosterName(2))
	assert.Equal(t, "ROSTER_TYPE_EXTENDED_LEAVE", RosterName(3))
	assert.Equal(t, "99", RosterName(99))
}

func TestRosterValueDefaults(t *testing.T) {
	Load()
	v, ok := RosterValue("ROSTER_TYPE_ACTIVE")
	assert.True(t, ok)
	assert.Equal(t, int32(1), v)

	_, ok = RosterValue("UNKNOWN")
	assert.False(t, ok)
}

func TestRosterIsCataloged(t *testing.T) {
	Load()
	assert.True(t, RosterIsCataloged(1))
	assert.False(t, RosterIsCataloged(99))
}

func TestAORForDefaults(t *testing.T) {
	Load()
	assert.Equal(t, "HHQ", AORFor("Regimental HQ"))
	assert.Equal(t, "HHQ", AORFor("Support Attachment"))
	assert.Equal(t, "RTC", AORFor("New Recruits"))
	assert.Equal(t, "Reserve", AORFor("7th Cavalry Reservists"))
	assert.Equal(t, "ELOA", AORFor("Extended Leave Of Absence"))
	assert.Equal(t, "HHQ", AORFor("Command Staff"))
	assert.Equal(t, "Alpha Company", AORFor("Alpha Company"))
}

func TestAwolActiveRostersDefault(t *testing.T) {
	Load()
	assert.Equal(t, []int{1, 2}, AwolActiveRosters())
}

func TestRankShortDefaults(t *testing.T) {
	Load()
	assert.Equal(t, "UNSPECIFIED", RankShort(0))
	assert.Equal(t, "COL", RankShort(6))
	assert.Equal(t, "PVT", RankShort(22))
	assert.Equal(t, "24", RankShort(24))
}

func TestRecordTypeNameDefaults(t *testing.T) {
	Load()
	assert.Equal(t, "RECORD_TYPE_UNSPECIFIED", RecordTypeName(0))
	assert.Equal(t, "RECORD_TYPE_PROMOTION", RecordTypeName(1))
	assert.Equal(t, "RECORD_TYPE_ELOA", RecordTypeName(8))
	assert.Equal(t, "99", RecordTypeName(99))
}

func TestRecordTypeIsCataloged(t *testing.T) {
	Load()
	assert.True(t, RecordTypeIsCataloged(1))
	assert.False(t, RecordTypeIsCataloged(99))
}

func TestIsUniformTriggerRecordType(t *testing.T) {
	Load()
	assert.True(t, IsUniformTriggerRecordType(1))
	assert.True(t, IsUniformTriggerRecordType(6))
	assert.False(t, IsUniformTriggerRecordType(2))
	assert.False(t, IsUniformTriggerRecordType(5))
}
