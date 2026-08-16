# DDI Roster Taxonomy Audit

> Branch: `ddi-api-data-conversion`  
> Scope: roster IDs, wire names, and API consumption. No code changes.

This audit focuses exclusively on the roster taxonomy after Phase 3.2. The roster system is now configurable through `internal/config` and the `ORG_ROSTER_NAMES` env var, but the defaults and the upstream `roster_id` values are still the 7CAV set.

## 1. Current Roster IDs and Meanings

The roster taxonomy is an `int32` enum in `types/roster.go` <ref_file file="/home/bobbyswagg/research/DDI-API/types/roster.go" />:

| `roster_id` | `types.RosterType` constant | Wire name (default) | 7CAV meaning |
| --- | --- | --- | --- |
| 0 | `RosterTypeUnspecified` | `ROSTER_TYPE_UNSPECIFIED` | invalid / null (rejected by handlers) |
| 1 | `RosterTypeCombat` | `ROSTER_TYPE_COMBAT` | active combat roster |
| 2 | `RosterTypeReserve` | `ROSTER_TYPE_RESERVE` | reserve roster |
| 3 | `RosterTypeEloa` | `ROSTER_TYPE_ELOA` | extended leave of absence |
| 4 | `RosterTypeWallOfHonor` | `ROSTER_TYPE_WALL_OF_HONOR` | wall of honor |
| 5 | `RosterTypeArlington` | `ROSTER_TYPE_ARLINGTON` | arlington / memorial roster |
| 6 | `RosterTypePastMembers` | `ROSTER_TYPE_PAST_MEMBERS` | past members |

These values are also the default catalog in `internal/config/orgmap.go` <ref_snippet file="/home/bobbyswagg/research/DDI-API/internal/config/orgmap.go" lines="108-114" /> and the `RosterType` schema in `openapi/openapi.yaml` <ref_snippet file="/home/bobbyswagg/research/DDI-API/openapi/openapi.yaml" lines="1017-1026" />.

## 2. Where Roster IDs Are Consumed

### 2.1 URL path parameters

Three endpoint families accept a `{roster}` path segment that is parsed by `ParseRosterType` <ref_file file="/home/bobbyswagg/research/DDI-API/types/roster.go" />:

* `GET /api/v1/roster/{roster}`
* `GET /api/v1/roster/{roster}/lite`
* `GET /api/v1/s1/uniforms/{roster}`

`ParseRosterType` accepts either the wire name or the raw `roster_id` number. The handler rejects the zero value with `cannot request null roster type`.

### 2.2 Datastore queries

`datastores/mysql.go` uses the parsed `RosterType` as a `roster_id` filter in three places:

* `FindRosterByType` <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="78-84" />
* `FindLiteRosterByType` <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="251-257" />
* `FindS1UniformsRosterByType` <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="337-344" />

In each case the SQL is `roster_id = uint(rosterType)` against `xf_nf_rosters_user`.

### 2.3 Wire emission

The `Roster` field on `Profile`, `LiteProfile`, and `S1UniformsProfile` is set from `profile.RosterId` <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="145-146" />. It is emitted as a name string via `RosterType.MarshalJSON`.

### 2.4 AWOL eligibility

`FindAwol` filters to `roster_id IN (config.AwolActiveRosters())`. The default is `1, 2` (combat and reserve) <ref_snippet file="/home/bobbyswagg/research/DDI-API/internal/config/orgmap.go" lines="166" />.

### 2.5 Test fixtures

`testdb/fixtures.sql` seeds the following `roster_id` members <ref_snippet file="/home/bobbyswagg/research/DDI-API/testdb/fixtures.sql" lines="85-103" />:

| `relation_id` | `roster_id` | Notes |
| --- | --- | --- |
| 1, 105, 205, 310 | 1 | combat roster |
| 320, 340 | 2 | reserve roster |
| 330, 350 | 6 | past members / discharged |
| — | 3 | empty (ELOA) |

The fixture comments explicitly state `roster_id=3` (ELOA) is intentionally empty to exercise the empty-roster-not-an-error test.

### 2.6 Golden corpus

The contract/golden files for `roster` use `ROSTER_TYPE_COMBAT`, `ROSTER_TYPE_RESERVE`, and the numeric form `1` <ref_file file="/home/bobbyswagg/research/DDI-API/contract/goldens/roster" />. `types/roster_test.go` pins the name/number parsing behavior <ref_file file="/home/bobbyswagg/research/DDI-API/types/roster_test.go" />.

## 3. Which Names Need DDI Replacement

Everything that is 7CAV-branded in the roster **names** should be replaced for a DDI deployment:

| Current wire name | DDI candidate (example) |
| --- | --- |
| `ROSTER_TYPE_COMBAT` | `ROSTER_TYPE_ACTIVE` |
| `ROSTER_TYPE_RESERVE` | `ROSTER_TYPE_RESERVE` (if DDI keeps a reserve) |
| `ROSTER_TYPE_ELOA` | `ROSTER_TYPE_ELOA` or `ROSTER_TYPE_LEAVE` |
| `ROSTER_TYPE_WALL_OF_HONOR` | `ROSTER_TYPE_HONOR_ROLL` or `ROSTER_TYPE_WALL_OF_HONOR` |
| `ROSTER_TYPE_ARLINGTON` | `ROSTER_TYPE_MEMORIAL` |
| `ROSTER_TYPE_PAST_MEMBERS` | `ROSTER_TYPE_FORMER_MEMBERS` or `ROSTER_TYPE_PAST_MEMBERS` |

The `ROSTER_TYPE_` prefix is part of the wire contract and should remain. Only the suffixes should change.

The `xf_nf_rosters.title` column in the database can hold any DDI titles (e.g. `Active Duty`, `Reserves`, `Leave of Absence`). Those are pure data and do not affect the wire form except for human readability and any `xf_nf_rosters` catalog endpoint that might be added later.

## 4. Which IDs Should Remain Stable

The `roster_id` values `1`–`6` should be kept as the canonical numeric identifiers for as long as DDI does not have a compelling reason to change them. Reasons to keep them stable:

* **Path compatibility**: clients can call `/api/v1/roster/1`, `/api/v1/roster/2`, etc. Changing the IDs would break those numeric calls.
* **Database foreign keys**: `xf_nf_rosters_user.roster_id` already uses `1`–`6` in fixtures and production. Re-mapping would require a data migration.
* **AWOL filter**: `ORG_AWOL_ACTIVE_ROSTERS` can be set to the same `1,2` or any other DDI values. The numeric IDs matter more than the names.
* **Golden corpus**: the contract goldens use the numeric id `1` for some paths, so changing it would require re-basing the golden corpus.
* **OpenAPI `RosterType` schema**: the schema lists the numeric values `0`–`6`. Keeping the IDs preserves the documented contract.

If DDI needs a different set of rosters, the recommended approach is to add new IDs (e.g. `7`, `8`) or reassign the **names** via `ORG_ROSTER_NAMES`, rather than deleting or renumbering the existing ones.

## 5. Recommended DDI Roster Structure

### 5.1 Minimal change (recommended for Phase 3.3)

Keep the existing `roster_id` values (`1`–`6`) and only change the wire names via `ORG_ROSTER_NAMES`:

```json
{
  "1": "ROSTER_TYPE_ACTIVE",
  "2": "ROSTER_TYPE_RESERVE",
  "3": "ROSTER_TYPE_ELOA",
  "4": "ROSTER_TYPE_HONOR_ROLL",
  "5": "ROSTER_TYPE_MEMORIAL",
  "6": "ROSTER_TYPE_FORMER_MEMBERS"
}
```

This satisfies:

* No database schema change
* No API route change
* No response schema change — only the `roster` string value changes
* No fixture `roster_id` renumbering
* No golden corpus renumbering

The `types/roster.go` constants (`RosterTypeCombat` etc.) and the `internal/config` default would still say `ROSTER_TYPE_*` in code, but runtime env overrides the catalog. A later cleanup phase can rename the Go constants if desired.

### 5.2 If DDI needs different IDs

If DDI cannot reuse `1`–`6`:

1. Add a **new catalog** in `types/roster.go` with DDI constants and `internal/config` defaults.
2. Update `xf_nf_rosters` and `xf_nf_rosters_user.roster_id` rows.
3. Update `ORG_AWOL_ACTIVE_ROSTERS` to the new active-roster IDs.
4. Update `testdb/fixtures.sql` and `contract/goldens/*`.
5. Update the `RosterType` schema in `openapi/openapi.yaml`.

This is more invasive and should only be done after the minimal name-change path is validated.

### 5.3 AWOL policy

`ORG_AWOL_ACTIVE_ROSTERS` should be set to whatever roster IDs represent active DDI members. If DDI keeps `1` and `2` as active and reserve, no change is needed.

## 6. Files to Touch in a Future Implementation Phase

If the decision is to rename only:

* `docker-compose.yaml` / production env: set `ORG_ROSTER_NAMES`.
* `openapi/openapi.yaml` `RosterType` enum: update the documented name list to match the env (or regenerate the spec).

If the decision is to renumber or restructure:

* `types/roster.go` constants
* `internal/config/orgmap.go` defaults
* `testdb/fixtures.sql` `roster_id` values
* `contract/goldens/roster/*`
* `datastores/mysql.go` AWOL filter (via `ORG_AWOL_ACTIVE_ROSTERS`)
* `openapi/openapi.yaml` `RosterType` schema

## 7. What Not to Change

Per the constraints:

* `xf_nf_rosters` and `xf_nf_rosters_user` table schemas
* `/api/v1/roster/{roster}` path templates
* `Roster` JSON field type and placement
* Rank/award/record taxonomy (out of scope)
* Handler and datastore query structure