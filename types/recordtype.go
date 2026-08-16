package types

import (
	"strconv"

	"github.com/BobbySwaggTV/DDI-API/internal/config"
)

// RecordType enumerates the categories of service-record entries. The
// numeric values match the proto enum being retired at Phase 4 exactly; the
// wire form is the NAME string (enum convention — see RosterType, the worked
// example). The name catalog now lives in internal/config so DDI can re-label
// record types without changing the numeric ids.
type RecordType int32

const (
	RecordTypeUnspecified  RecordType = 0
	RecordTypePromotion    RecordType = 1
	RecordTypeOperation    RecordType = 2
	RecordTypeTransfer     RecordType = 3
	RecordTypeDisciplinary RecordType = 4
	RecordTypeDischarge    RecordType = 5
	RecordTypeAssignment   RecordType = 6
	RecordTypeNameChange   RecordType = 7
	RecordTypeEloa         RecordType = 8
	RecordTypeGraduation   RecordType = 9
)

// String returns the wire name, or the bare decimal number for values the
// catalog has no name for — mirroring the generated proto String() (the
// RosterType worked example), so enum values interpolate identically into
// error-message strings and logs.
func (rt RecordType) String() string {
	return config.RecordTypeName(int32(rt))
}

// MarshalJSON emits the enum name as a JSON string; values without a name
// emit the bare number (protojson behavior for unknown enum values).
func (rt RecordType) MarshalJSON() ([]byte, error) {
	if config.RecordTypeIsCataloged(int32(rt)) {
		return strconv.AppendQuote(nil, rt.String()), nil
	}
	return strconv.AppendInt(nil, int64(rt), 10), nil
}
