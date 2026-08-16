# DDI Record Taxonomy Audit

> Branch: `ddi-api-data-conversion`  
> Scope: record types, wire names, and uniform-update policy. No code changes.

This audit documents the service-record taxonomy, where the IDs are consumed, which values are 7CAV-specific, and how to move the record naming and meaning under DDI control without changing the database schema or response structures.

## 1. Current Record Types and Meanings

`types/recordtype.go` hardcodes the record-type enum. The numeric values are the `record_type_id` foreign key in `xf_nf_rosters_service_record` <ref_file file="/home/bobbyswagg/research/DDI-API/types/recordtype.go" />:

| `record_type_id` | Wire name | 7CAV meaning |
| --- | --- | --- |
| 0 | `RECORD_TYPE_UNSPECIFIED` | invalid / placeholder |
| 1 | `RECORD_TYPE_PROMOTION` | promotion |
| 2 | `RECORD_TYPE_OPERATION` | operation |
| 3 | `RECORD_TYPE_TRANSFER` | transfer |
| 4 | `RECORD_TYPE_DISCIPLINARY` | disciplinary action |
| 5 | `RECORD_TYPE_DISCHARGE` | discharge |
| 6 | `RECORD_TYPE_ASSIGNMENT` | assignment |
| 7 | `RECORD_TYPE_NAME_CHANGE` | name change |
| 8 | `RECORD_TYPE_ELOA` | extended leave of absence |
| 9 | `RECORD_TYPE_GRADUATION` | graduation |

The database catalog table `xf_nf_rosters_record_type` is not queried by the API; the `title` column in that table is effectively dead data from the code's perspective. The wire value is taken entirely from `types/recordtype.go`.

`testdb/fixtures.sql` seeds the following service records <ref_snippet file="/home/bobbyswagg/research/DDI-API/testdb/fixtures.sql" lines="112-120" />:

| `record_id` | `relation_id` | `details` | `record_type_id` |
| --- | --- | --- | --- |
| 1001 | 1 | `Promoted to Sergeant Major` | 1 (PROMOTION) |
| 1002 | 1 | `Assigned to Squad Leader` | 2 (OPERATION) |
| 1003 | 205 | `Graduated basic training` | 5 (DISCHARGE in fixtures; note it is a deliberate mismatch with the text) |
| 1004 | 205 | `Promoted to Specialist` | 1 (PROMOTION) |
| 1005 | 310 | `Enlisted` | 5 (DISCHARGE) |
| 1006 | 320 | `Transferred to reserve` | 2 (OPERATION) |
| 1007 | 330 | `Honorably discharged` | 3 (TRANSFER) |

The fixture comments explicitly state the `details` text is decorative and the `record_type_id` is load-bearing for tests.

## 2. Where Record Types Are Consumed

### 2.1 Wire output

`Record` appears in the `records` array of `Profile` and `LiteProfile`:

```go
type Record struct {
    RecordDetails string     `json:"recordDetails"`
    RecordType    RecordType `json:"recordType"`
    RecordDate    string     `json:"recordDate"`
    RecordUid     uint64     `json:"recordUid,string"`
}
```

<ref_snippet file="/home/bobbyswagg/research/DDI-API/types/profiles.go" lines="52-58" />

`RecordType` is serialized by `MarshalJSON` as the name string and has a `String()` method used in error messages and logs <ref_file file="/home/bobbyswagg/research/DDI-API/types/recordtype.go" />.

### 2.2 Datastore mapping

`datastores/mysql.go` `collectRecords` maps the raw `record_type_id` to `types.RecordType` for the wire <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="209-218" />.

### 2.3 Uniform update trigger logic

`datastores/mysql.go` `getUniformUpdateTriggerDate` hardcodes a set of record types that reset the uniform-update clock <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="429-436" />:

```go
relevantRecordTypes := map[types.RecordType]bool{
    types.RecordTypePromotion:  true,
    types.RecordTypeAssignment: true,
    types.RecordTypeEloa:       true,
    types.RecordTypeNameChange: true,
    types.RecordTypeGraduation: true,
}
```

This is a 7CAV policy decision. DDI may have a different set of events that should trigger an S1 uniform recalculation.

### 2.4 Tests and goldens

`types/recordtype_test.go` pins the wire names for the enum <ref_file file="/home/bobbyswagg/research/DDI-API/types/recordtype_test.go" />. `datastores/profiles_harness_test.go` asserts a `PROMOTION` record for relation 205 <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/profiles_harness_test.go" lines="112-113" />. The contract golden files for `roster` and `milpacs/profile` include `recordType` fields.

## 3. Which Values Are 7CAV-Specific

The `RECORD_TYPE_` wire names themselves are not strongly 7CAV-branded except for `RECORD_TYPE_ELOA` (which mirrors the 7CAV `ROSTER_TYPE_ELOA`). However, the entire set is U.S. Army / military terminology:

* `PROMOTION`, `DISCHARGE`, `ASSIGNMENT`, `GRADUATION` — generic but military flavored
* `OPERATION` — 7CAV framing for an assignment/operation record
* `TRANSFER` — generic
* `DISCIPLINARY` — generic but military flavored
* `NAME_CHANGE` — generic
* `ELOA` — 7CAV-specific acronym

The most 7CAV-specific assumption is the **uniform update trigger set**: promotion, assignment, ELOA, name change, graduation. That set encodes the 7CAV uniform-update business rule.

## 4. Which IDs/Enums Should Remain Stable

The `record_type_id` values `1`–`9` should be kept stable for the same reasons as `roster_id`:

* They are stored in `xf_nf_rosters_service_record.record_type_id`.
* They are embedded in the contract golden corpus.
* They are used in the test fixture logic (e.g. record 1002 is type 2 to test the `OPERATION` wire name and the non-relevant record filter).
* Changing them would require a data migration and a golden rebase.

The zero value `0` (`RECORD_TYPE_UNSPECIFIED`) is a placeholder and is not emitted in real records.

## 5. Recommended DDI Record Taxonomy Migration Strategy

### 5.1 Make record wire names configurable

Mirror the `types.RosterType` / `internal/config` pattern:

1. Move the `recordTypeNames` catalog from `types/recordtype.go` to `internal/config` (or add a `RecordTypeNames` map).
2. Add a `RECORD_TYPE_NAMES` env var: JSON `record_type_id` → wire name.
3. Have `RecordType.String()` and `RecordType.MarshalJSON()` read from the config catalog.
4. Default to the current 7CAV names so existing tests and fixtures pass.

DDI can then override without touching the database or API shapes:

```json
{
  "1": "RECORD_TYPE_PROMOTION",
  "2": "RECORD_TYPE_OPERATION",
  "3": "RECORD_TYPE_TRANSFER",
  "4": "RECORD_TYPE_DISCIPLINARY",
  "5": "RECORD_TYPE_DISCHARGE",
  "6": "RECORD_TYPE_ASSIGNMENT",
  "7": "RECORD_TYPE_NAME_CHANGE",
  "8": "RECORD_TYPE_LEAVE",
  "9": "RECORD_TYPE_GRADUATION"
}
```

### 5.2 Make the uniform trigger set configurable

1. Add a `UNIFORM_TRIGGER_RECORD_TYPES` env var: comma-separated `record_type_id` values.
2. Replace the hardcoded `relevantRecordTypes` map in `datastores/mysql.go` `getUniformUpdateTriggerDate` with a call to `config.UniformTriggerRecordTypes()`.
3. Default to `1,6,7,8,9` (promotion, assignment, name change, ELOA/leave, graduation) so current behavior is preserved.

This makes the DDI uniform-update policy data-driven.

### 5.3 Keep `record_id` values stable

Do not renumber `record_type_id` `1`–`9`. DDI can change the wire names and the trigger set without changing the foreign keys.

### 5.4 Record details are pure data

`RecordDetails` comes from `xf_nf_rosters_service_record.details`. DDI can update the text directly; no code changes are needed. The fixture data's decorative mismatches should be reviewed and corrected when DDI re-seeds the database.

### 5.5 Update fixtures and goldens only after renaming

When `RECORD_TYPE_NAMES` is set to DDI wire names:

* `types/recordtype_test.go` expected strings
* `types/profiles_test.go` and `datastores/profiles_harness_test.go` assertions
* `contract/goldens/*` `recordType` fields
* `openapi/openapi.yaml` `RecordType` enum list

The `record_type_id` values inside those records stay the same.

## 6. Files to Touch in a Future Implementation Phase

If adding the configuration layer:

* `internal/config/orgmap.go` — add `RecordTypeNames` and `UniformTriggerRecordTypes`
* `types/recordtype.go` — `String()` and `MarshalJSON()` use the config catalog
* `datastores/mysql.go` `getUniformUpdateTriggerDate` — replace hardcoded set with `config.UniformTriggerRecordTypes()`
* `cmd/root.go` — document `RECORD_TYPE_NAMES` and `UNIFORM_TRIGGER_RECORD_TYPES`
* `docker-compose.yaml` — add env examples

If re-seeding data:

* `testdb/fixtures.sql` `xf_nf_rosters_service_record` rows (details text only, keep ids)
* `xf_nf_rosters_record_type` `title` column (optional; code does not read it)

## 7. What Not to Change

Per the constraints:

* `xf_nf_rosters_service_record` and `xf_nf_rosters_record_type` table schemas
* `record_type_id` numeric values
* `Record` JSON field names
* `recordType` wire type (still a string enum)
* Any endpoint paths