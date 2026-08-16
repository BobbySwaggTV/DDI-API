# DDI Organization and AOR Mapping Audit

> Branch: `ddi-api-data-conversion`  
> Scope: organization-specific data and mapping logic. No code changes.

This document lists the 7CAV organization assumptions embedded in the code, explains what each roster and rank/record ID means, and recommends how to replace them with DDI values without changing the database schema, response structures, or endpoints.

## 1. Hardcoded 7CAV Organization Assumptions

### 1.1 Roster taxonomy

`types/roster.go` hardcodes the roster enum. The numeric values are the `roster_id` foreign key in `xf_nf_rosters_user` and `xf_nf_rosters`:

| `roster_id` | Wire name | 7CAV meaning |
| --- | --- | --- |
| 0 | `ROSTER_TYPE_UNSPECIFIED` | invalid / null roster (rejected by handlers) |
| 1 | `ROSTER_TYPE_COMBAT` | active combat roster |
| 2 | `ROSTER_TYPE_RESERVE` | reserve roster |
| 3 | `ROSTER_TYPE_ELOA` | extended leave of absence |
| 4 | `ROSTER_TYPE_WALL_OF_HONOR` | wall of honor |
| 5 | `ROSTER_TYPE_ARLINGTON` | arlington / memorial roster |
| 6 | `ROSTER_TYPE_PAST_MEMBERS` | past members |

<ref_file file="/home/bobbyswagg/research/DDI-API/types/roster.go" />

The `ParseRosterType` function accepts either the wire name or the raw number. The roster lookup queries use `roster_id = uint(rosterType)` in `datastores/mysql.go` <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="80-84" />.

### 1.2 AWOL active-roster filter

`datastores/mysql.go` `FindAwol` only considers `roster_id IN (1, 2)` and a 7-day cutoff <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="736-737" />. This is a 7CAV policy assumption: active members live on rosters 1 or 2, and AWOL is 7 days without a post.

### 1.3 Position-group AOR mapping

`datastores/mysql.go` `getPositionGroup` hardcodes 7CAV unit names to an area-of-responsibility string <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="457-475" />:

```go
"Regimental HQ", "Support Attachment" -> "HHQ"
"New Recruits"                        -> "RTC"
"7th Cavalry Reservists"              -> "Reserve"
"Extended Leave Of Absence"           -> "ELOA"
title contains "Command"              -> "HHQ"
otherwise                             -> title
```

This feeds the `areaOfResponsibility` field in `S1UniformsProfile`.

### 1.4 Rank taxonomy

`types/ranks.go` hardcodes a U.S. Army rank enum. The values are the `rank_id` foreign key in `xf_nf_rosters_rank`:

| `rank_id` | RankShort | Full title |
| --- | --- | --- |
| 1 | `GOA` | General of the Army |
| 2 | `GEN` | General |
| 3 | `LTG` | Lieutenant General |
| 4 | `MG` | Major General |
| 5 | `BG` | Brigadier General |
| 6 | `COL` | Colonel |
| ... | ... | ... |
| 22 | `PVT` | Private |
| 23 | `RCT` | Recruit |
| 24, 25 | (unassigned) | fallback to decimal string |
| 26-31 | `CW5`..`AR` | warrant and auxiliary ranks |

The short name is derived by stripping `RANK_TYPE_` from the enum name <ref_snippet file="/home/bobbyswagg/research/DDI-API/types/ranks.go" lines="122-124" />. The `datastores/mysql.go` mapper stamps this short name onto every profile and rank catalog entry <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="136-139" />.

### 1.5 Service-record taxonomy

`types/recordtype.go` hardcodes 7CAV record categories. The values are the `record_type_id` foreign key in `xf_nf_rosters_service_record`:

| `record_type_id` | Wire name | Meaning |
| --- | --- | --- |
| 0 | `RECORD_TYPE_UNSPECIFIED` | invalid / placeholder |
| 1 | `RECORD_TYPE_PROMOTION` | promotion |
| 2 | `RECORD_TYPE_OPERATION` | operation |
| 3 | `RECORD_TYPE_TRANSFER` | transfer |
| 4 | `RECORD_TYPE_DISCIPLINARY` | disciplinary action |
| 5 | `RECORD_TYPE_DISCHARGE` | discharge |
| 6 | `RECORD_TYPE_ASSIGNMENT` | assignment |
| 7 | `RECORD_TYPE_NAME_CHANGE` | name change |
| 8 | `RECORD_TYPE_ELOA` | ELOA |
| 9 | `RECORD_TYPE_GRADUATION` | graduation |

<ref_file file="/home/bobbyswagg/research/DDI-API/types/recordtype.go" />

### 1.6 Uniform-update trigger record types

`datastores/mysql.go` `getUniformUpdateTriggerDate` decides which record types count as a "relevant" update for the S1 uniform trigger date <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="428-435" />:

```go
relevantRecordTypes := map[types.RecordType]bool{
    types.RecordTypePromotion:  true,
    types.RecordTypeAssignment: true,
    types.RecordTypeEloa:       true,
    types.RecordTypeNameChange: true,
    types.RecordTypeGraduation: true,
}
```

This is a 7CAV policy: only promotions, assignments, ELOA, name changes, and graduations reset the uniform-update clock.

### 1.7 Fixture data

`testdb/fixtures.sql` seeds 7CAV-specific position groups and titles <ref_snippet file="/home/bobbyswagg/research/DDI-API/testdb/fixtures.sql" lines="52-67" />:

| `position_group_id` | title |
| --- | --- |
| 1 | `Regimental HQ` |
| 2 | `New Recruits` |
| 3 | `Alpha Company` |
| 4 | `Extended Leave Of Absence` |

| `roster_id` | fixture members |
| --- | --- |
| 1 | combat roster |
| 2 | reserve roster |
| 6 | past members |

The rank catalog also uses 7CAV sample titles (`Sergeant Major`, `Private`, etc.).

### 1.8 Connected accounts and custom fields

* Discord provider string: `nfDiscord` <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="109-110" />.
* Custom field JSON keys: `realName`, `joinDate`, `promoDate`, `mos`, `consoleGamertag` <ref_snippet file="/home/bobbyswagg/research/DDI-API/milpacs/profiles.go" lines="61-67" />.

These are not schema; they are data conventions the code assumes for connected-account and custom-field lookups.

## 2. Current Meaning of Each Roster ID

| ID | Enum | Usage |
| --- | --- | --- |
| 0 | `ROSTER_TYPE_UNSPECIFIED` | Parses to the zero value; handlers reject it ("cannot request null roster type") |
| 1 | `ROSTER_TYPE_COMBAT` | Active combat roster; included in AWOL filter |
| 2 | `ROSTER_TYPE_RESERVE` | Reserve roster; included in AWOL filter |
| 3 | `ROSTER_TYPE_ELOA` | Extended leave; intentionally empty in fixtures |
| 4 | `ROSTER_TYPE_WALL_OF_HONOR` | Wall of honor roster |
| 5 | `ROSTER_TYPE_ARLINGTON` | Arlington roster |
| 6 | `ROSTER_TYPE_PAST_MEMBERS` | Past members; used in fixture for a discharged member |

The API accepts `/api/v1/roster/ROSTER_TYPE_COMBAT` or `/api/v1/roster/1`. The handler calls `ParseRosterType`, which validates the value exists in the `rosterTypeNames` catalog. The wire response always emits the name string.

## 3. Current AOR / Position Mapping Logic

### 3.1 Where AOR is computed

Only the S1 uniform response uses `areaOfResponsibility`. The chain is:

1. `GET /api/v1/s1/uniforms/{roster}` -> `FindS1UniformsRosterByType`
2. `generateS1UniformsProtoProfile`
3. `getPositionGroup(profile)`

`getPositionGroup` reads `profile.Primary.Group.Title` and applies the switch described in section 1.3.

### 3.2 What is not mapped

* Full `Profile` and `LiteProfile` do **not** include `areaOfResponsibility`; they expose the raw `primary` position object and `secondaries`.
* Position group and position titles flow through to the wire unchanged, except for the S1 AOR field.

### 3.3 Position divider rows

Position `50` in the fixture (`---- Staff ----`) is a visual divider. `FindAllPositionGroups` filters out rows whose title starts with `----` <ref_file file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" />. This is presentation logic, not organization identity.

## 4. Recommended DDI Replacement Approach

### 4.1 Roster taxonomy

1. Replace `types/roster.go` `rosterTypeNames`/`rosterTypeValues` with DDI roster names and the same or new `roster_id` values.
2. If DDI uses different IDs, update the `roster_id` values in `xf_nf_rosters` and `xf_nf_rosters_user` fixture data.
3. Update `datastores/mysql.go` `FindAwol` `roster_id IN (?)` to the DDI rosters that represent active status.
4. If the 7-day AWOL cutoff is different, introduce an `AWOL_CUTOFF_DAYS` env var and use it instead of `time.Now().AddDate(0, 0, -7)`.

### 4.2 AOR / position mapping

The cleanest no-schema-change path is to make the AOR map configurable or data-driven:

* **Option A (recommended for Phase 3/4)**: read an env var such as `POSITION_GROUP_AOR='{"Regimental HQ":"HHQ","New Recruits":"RTC"}'` and replace `getPositionGroup` with a map lookup. This keeps the API rebuildable for any DDI reorganization.
* **Option B**: simply rewrite the `switch` in `getPositionGroup` to DDI unit titles. This is faster but hardcodes DDI names in the binary.

For either option, the position group titles in `xf_nf_rosters_position_group` should be updated to DDI names, and `getPositionGroup` must match those titles.

### 4.3 Rank taxonomy

1. Replace `types/ranks.go` `rankTypeNames` and the `RankType` constants with DDI rank IDs and names.
2. Change `RankShort()` to strip a DDI prefix (e.g. `DDI_RANK_`) or replace the entire derivation with a static map if DDI rank short names do not derive from long names.
3. Update `xf_nf_rosters_rank` fixture data to use DDI rank IDs and titles.
4. If the rank catalog should not assume a `rank_id` → `RankShort` derivation, add a `rank_short` column or read the short name from the `title` column — but since we cannot change schema, derive it in code.

### 4.4 Record taxonomy

1. Replace `types/recordtype.go` `recordTypeNames` with DDI record categories.
2. Update `datastores/mysql.go` `getUniformUpdateTriggerDate` `relevantRecordTypes` set to the DDI record types that should trigger a uniform update.
3. Update `xf_nf_rosters_record_type` fixture data.

### 4.5 Connected accounts and custom fields

1. Replace or configure the Discord provider string (`nfDiscord`) if DDI uses a different connected-account provider. If DDI keeps Discord, make the provider name a viper value.
2. Replace or configure custom field JSON keys if DDI stores `realName`, `joinDate`, `mos`, etc. under different keys.

### 4.6 Fixtures

`testdb/fixtures.sql` must be reseeded with DDI values for:

* `xf_nf_rosters` titles and IDs
* `xf_nf_rosters_user.roster_id`
* `xf_nf_rosters_rank` IDs/titles
* `xf_nf_rosters_position_group` titles
* `xf_nf_rosters_position` titles
* `xf_nf_rosters_record_type` titles/IDs
* `xf_nf_rosters_service_record.record_type_id`

The `contract/goldens/*` should then be regenerated by running the contract suite against the new fixtures.

## 5. Implementation Order

1. **Roster enum and AWOL filter** (`types/roster.go`, `datastores/mysql.go`)
2. **AOR mapping** (`datastores/mysql.go` `getPositionGroup`)
3. **Rank enum and short-name logic** (`types/ranks.go`, `datastores/mysql.go`)
4. **Record enum and uniform-trigger logic** (`types/recordtype.go`, `datastores/mysql.go`)
5. **Connected-account provider and custom-field keys** (`datastores/mysql.go`, `milpacs/profiles.go`)
6. **Re-seed fixtures and goldens** (`testdb/fixtures.sql`, `contract/goldens/*`)
7. **Run `go test ./...` and `make test-integration`**

At each step, the database schema stays the same, the API endpoints stay the same, and the wire types stay the same. Only the data values and the mapping functions change.