# DDI Data Conversion Plan

> Branch: `ddi-api-data-conversion`  
> Scope: data, identity, and asset migration. No code changes.

This plan identifies the 7CAV-specific data and configuration that must be replaced for DDI, and the order in which to do it without changing the database schema, the API surface, or the test endpoints.

## 1. API Key System

### 1.1 All `xf_cav7_api_key` references

The API authenticates against the Cav7 ApiKeyManager tables. These are the only places the table names are used:

* `datastores/mysql.go` — `ValidateApiKey` and the `last_used_date` bump query <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="761-804" />.
* `testdb/schema.sql` — schema definition <ref_snippet file="/home/bobbyswagg/research/DDI-API/testdb/schema.sql" lines="400-423" />.
* `testdb/fixtures.sql` — test fixture inserts <ref_snippet file="/home/bobbyswagg/research/DDI-API/testdb/fixtures.sql" lines="299-318" />.
* `testdb/testdb.go` — test constants (`ActiveAPIKey`, `RevokedAPIKey`) <ref_snippet file="/home/bobbyswagg/research/DDI-API/testdb/testdb.go" lines="40-46" />.
* `testdb/testdb_test.go` — schema-coverage assertions <ref_snippet file="/home/bobbyswagg/research/DDI-API/testdb/testdb_test.go" lines="96-150" />.
* `datastores/auth_harness_test.go` — `last_used_date` column tests.

### 1.2 Safe conversion strategy

The instruction is to **not rename the tables**. The `xf_cav7_api_key*` tables are an internal implementation detail — the public API only exposes `Authorization: Bearer <token>` and scopes. Therefore the safest approach is:

1. **Leave the table names as-is in the first data-conversion pass.** The client never sees the table names; the 7CAV branding is internal to the database.
2. **If internal branding must be removed**, create a new DDI table set (e.g. `xf_ddi_api_key`, `xf_ddi_api_key_scope`, `xf_ddi_api_key_scope_def`) and point `ValidateApiKey` to it. Do not drop or rename `xf_cav7_api_key*` — keep it as a historical backstop.
3. **Scope names** (`read`, `read:tickets`) are generic and can stay.
4. **Validation logic** (`UNHEX(SHA2(?, 256))`) is also generic and can stay.

This avoids touching the upstream forum schema while giving DDI a clean path to its own key store later.

## 2. Organization Identity

### 2.1 Hardcoded roster IDs

`types.RosterType` in `types/roster.go` hardcodes the roster enum:

| Value | Name | Meaning | Used by |
| --- | --- | --- | --- |
| 1 | `ROSTER_TYPE_COMBAT` | active combat roster | `FindRosterByType`, `FindLiteRosterByType`, `FindS1UniformsRosterByType` |
| 2 | `ROSTER_TYPE_RESERVE` | reserve roster | same as above, plus AWOL filter |
| 3 | `ROSTER_TYPE_ELOA` | extended leave | same as above |
| 4 | `ROSTER_TYPE_WALL_OF_HONOR` | wall of honor | same as above |
| 5 | `ROSTER_TYPE_ARLINGTON` | arlington | same as above |
| 6 | `ROSTER_TYPE_PAST_MEMBERS` | past members | same as above |

The route accepts either the enum name or the raw integer. The `FindAwol` query filters to `roster_id IN (1, 2)` <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="736" />.

These are the **7CAV organization IDs**. DDI must define its own roster taxonomy. The schema itself is not constrained; the `roster_id` column can hold DDI values. The code needs:

* Replace or extend `types.RosterType`/`rosterTypeNames`.
* Update the `FindAwol` active-roster filter if DDI AWOL policy differs.
* Reseed `xf_nf_rosters` and `xf_nf_rosters_user.roster_id` with DDI values.

### 2.2 Position / unit IDs and AOR mappings

The fixtures insert these position groups <ref_snippet file="/home/bobbyswagg/research/DDI-API/testdb/fixtures.sql" lines="52-56" />:

| position_group_id | title | AOR mapping |
| --- | --- | --- |
| 1 | `Regimental HQ` | `HHQ` |
| 2 | `New Recruits` | `RTC` |
| 3 | `Alpha Company` | (title itself) |
| 4 | `Extended Leave Of Absence` | `ELOA` |

`datastores/mysql.go` `getPositionGroup` hardcodes the AOR (area-of-responsibility) logic <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="457-475" />:

```go
"Regimental HQ", "Support Attachment" -> "HHQ"
"New Recruits"                        -> "RTC"
"7th Cavalry Reservists"              -> "Reserve"
"Extended Leave Of Absence"           -> "ELOA"
title contains "Command"              -> "HHQ"
otherwise                             -> title
```

These are 7CAV unit names and AOR values. For DDI:

* Keep the `position_group_id` values if they align with DDI units, or replace them.
* Replace the `title` strings in `xf_nf_rosters_position_group`.
* Rewrite `getPositionGroup` for DDI organization structure.

### 2.3 Connected-account and custom-field identifiers

* Discord provider: `provider = 'nfDiscord'` in `datastores/mysql.go` <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="109-110" /> and `extractDiscordID` <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="162-169" />.
* Custom field JSON keys: `realName`, `joinDate`, `promoDate`, `mos`, `consoleGamertag` <ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/profiles.go" />.
* Field value join uses `field_id = 'consoleGamertag'` for gamertag lookup <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="815" />.

These are not schema; they are data conventions that DDI can either keep or remap.

## 3. Asset Ownership

### 3.1 All `7cav.us` URLs

Production assets are built in `milpacs/*.go`:

* Rank images: `https://7cav.us/data/roster_ranks/{group}/{rankId}.jpg?{rankImage}` in `milpacs/ranks.go` <ref_snippet file="/home/bobbyswagg/research/DDI-API/milpacs/ranks.go" lines="38-41" />.
* Uniform images: `https://7cav.us/data/roster_uniforms/{group}/{relationId}.jpg` in `milpacs/profiles.go` <ref_snippet file="/home/bobbyswagg/research/DDI-API/milpacs/profiles.go" lines="56-59" />.
* Award images: `https://7cav.us/data/roster_awards/{group}/{awardId}.jpg?{awardImage}` in `milpacs/awards.go` <ref_snippet file="/home/bobbyswagg/research/DDI-API/milpacs/awards.go" lines="39-42" />.
* Ticket `forumUrl` is already derived from `FORUM_BASE_URL` <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/tickets.go" lines="265-267" />.

Test and golden files also contain `7cav.us` example URLs; those will be regenerated with DDI fixtures.

### 3.2 Replacement strategy

The cleanest no-schema change is to externalize the asset base URL:

1. Add an `ASSET_BASE_URL` env var (read via `viper` like `FORUM_BASE_URL`).
2. Replace the hardcoded `https://7cav.us/data` string in the three `ImageURL()` / `UniformUrl()` methods with the configured base.
3. Keep the same path template structure (`/roster_ranks/{group}/{id}.jpg?{cache}` etc.) so existing asset pipelines and CDN rules work without additional code changes.
4. In DDI production, set `ASSET_BASE_URL=https://assets.daggerdefenseindustries.com/data` or similar.
5. The database still stores `rank_image`, `award_image` cache-busters and `uniform_date` timestamps; only the base URL changes in code.

### 3.3 Future DDI asset structure

Recommended structure mirrors the current one:

```
{ASSET_BASE_URL}/roster_ranks/{group}/{rankId}.jpg?{rankImage}
{ASSET_BASE_URL}/roster_uniforms/{group}/{relationId}.jpg
{ASSET_BASE_URL}/roster_awards/{group}/{awardId}.jpg?{awardImage}
```

This minimizes code and migration risk. If DDI later wants a different taxonomy, it can be introduced after the base-URL change.

## 4. Data Models

### 4.1 Ranks

* **What can remain**: `xf_nf_rosters_rank` columns, the GORM `milpacs.Rank` model, the wire `Rank`/`RankExpanded` types, and the image URL template.
* **What needs DDI mapping**: `types.RankType` enum (`RANK_TYPE_GOA` through `RANK_TYPE_AR`), the `RankShort()` logic that strips `RANK_TYPE_`, and the actual `rank_id`/`title` rows in `xf_nf_rosters_rank`.

DDI can reuse the `rank_id` column values or replace them; the code only cares about the integer and the `title`/`rank_image` columns.

### 4.2 Positions

* **What can remain**: `xf_nf_rosters_position`, `xf_nf_rosters_position_group` schemas and GORM models; `Position`/`PositionExpanded` wire types.
* **What needs DDI mapping**: `position_title` and `position_group.title` values; the `getPositionGroup` AOR switch in `datastores/mysql.go`; and possibly `possible_secondary` flags.

### 4.3 Rosters

* **What can remain**: `xf_nf_rosters` and `xf_nf_rosters_user` schemas, `Roster`/`LiteRoster`/`S1UniformsRoster` wire types, route patterns.
* **What needs DDI mapping**: `types.RosterType` names/values, `xf_nf_rosters.title` rows, and `xf_nf_rosters_user.roster_id` foreign-key values.

### 4.4 Records

* **What can remain**: `xf_nf_rosters_service_record` and `xf_nf_rosters_record_type` schemas, `Record` wire type.
* **What needs DDI mapping**: `types.RecordType` enum (`RECORD_TYPE_PROMOTION` etc.), `xf_nf_rosters_record_type.title` values, and the `record_type_id` set used for the uniform update trigger (`PROMOTION`, `ASSIGNMENT`, `ELOA`, `NAME_CHANGE`, `GRADUATION`) <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="428-435" />.

### 4.5 What is truly generic

The following do **not** need DDI mapping unless the schema changes:

* The `Profile`/`LiteProfile`/`S1UniformsProfile` JSON shapes and field names.
* Bearer-token auth flow and scopes.
* Cache-Control max-age values.
* The `net/http` route and middleware stack.
* The query-binding and wire-convention rules.

## 5. Implementation Order

This ordering keeps the API buildable and testable after each step.

1. **Externalize asset URLs** (`milpacs/ranks.go`, `milpacs/awards.go`, `milpacs/profiles.go`)
   * Add `ASSET_BASE_URL` env var.
   * No schema, response, or endpoint change.
   * Update tests and goldens when fixtures are reseeded.

2. **Replace AOR mapping** (`datastores/mysql.go` `getPositionGroup`)
   * Make it data-driven or re-code for DDI position-group titles.
   * Update `testdb/fixtures.sql` with DDI position groups.

3. **Roster taxonomy** (`types/roster.go`, `datastores/mysql.go` `FindAwol` filter)
   * Replace `RosterType` enum and `rosterTypeNames` with DDI values.
   * Update `xf_nf_rosters` fixture data.

4. **Rank taxonomy** (`types/ranks.go`, `milpacs/ranks.go` image URL)
   * Replace `RankType` and `rankTypeNames` with DDI ranks.
   * Update `xf_nf_rosters_rank` fixture data and golden images.

5. **Record taxonomy** (`types/recordtype.go`, `datastores/mysql.go` `getUniformUpdateTriggerDate`)
   * Replace `RecordType` with DDI record categories.
   * Update the uniform-update-trigger record-type set.

6. **Connected accounts and custom fields**
   * Replace or configure `nfDiscord` provider string.
   * Replace or configure `consoleGamertag` field id and custom field JSON keys.

7. **API key store (optional)**
   * Create `xf_ddi_api_key*` tables and switch `datastores/mysql.go` to them.
   * Keep `xf_cav7_api_key*` in place for rollback.

8. **Regenerate integration fixtures and golden corpus**
   * `testdb/fixtures.sql` and `contract/goldens/*` must reflect DDI data, assets, and unit names.

9. **Run full verification**
   * `go test ./...`
   * `make test-integration` with a MariaDB harness.

## 6. What Not to Touch in This Phase

Per the constraints:

* Database table names — do not rename or drop `xf_cav7_api_key*`, `xf_nf_rosters*`, `xf_nf_tickets*`, etc.
* API response wire types and JSON field names.
* Public endpoint paths and HTTP methods.
* Test structure — only their data and assertions, not the test harness itself.