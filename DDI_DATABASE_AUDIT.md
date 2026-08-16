# DDI Database Audit

> Branch: `ddi-conversion`  
> Scope: database architecture, dependencies, and migration options. No code changes.

This document describes the database foundation of the API and the migration trade-offs for DDI. The API is a read-only GORM/MySQL layer; it does not create or migrate schema. The source-of-truth schema lives in the upstream XenForo + add-on database.

## 1. Database Architecture

### 1.1 Database engine

* **Engine**: MariaDB / MySQL (production is XenForo on MariaDB 11.5).
* **Driver**: `github.com/go-sql-driver/mysql` via `gorm.io/driver/mysql`.
* **Database name**: `xenforo` — hardcoded in the DSN in `servers/server.go`.
* **ORM**: GORM 1.31.2 over `database/sql`.
* **Read-only posture**: the service never runs DDL. Indexes and table definitions are managed by the upstream forum.

Connection string construction <ref_snippet file="/home/bobbyswagg/research/DDI-API/servers/server.go" lines="72-100" />:

```go
"%s:%s@tcp(%s:%s)/xenforo?charset=utf8mb4&parseTime=True&loc=Local"
```

### 1.2 Connection handling

Configuration is env-only:

* `DB_USERNAME`
* `DB_PASSWORD`
* `DB_HOST`
* `DB_PORT`
* `DB_MAX_OPEN_CONNS`
* `DB_MAX_IDLE_CONNS`
* `DB_CONN_MAX_LIFETIME`

`servers/pool.go` resolves these with conservative defaults designed for shared hosting with the forum:

| Setting | Default | Purpose |
| --- | --- | --- |
| `MaxOpenConns` | 25 | cap in-flight connections |
| `MaxIdleConns` | 25 | keep pool warm |
| `ConnMaxLifetime` | 30m | recycle before `wait_timeout` |

<ref_file file="/home/bobbyswagg/research/DDI-API/servers/pool.go" />

### 1.3 Required tables

The API reads from four logical groups of tables.

#### Core XenForo

| Table | Columns read | Why the API needs it |
| --- | --- | --- |
| `xf_user` | `user_id`, `username` | resolves forum identity for `Profile.User` |
| `xf_user_group` | `user_group_id`, `title` | `GET /api/v1/forum/groups` |
| `xf_user_connected_account` | `user_id`, `provider`, `provider_key` | Discord lookup (`provider = 'nfDiscord'`) and `discordId` field extraction |
| `xf_post` | `user_id`, `post_date`, `post_id` | last-forum-post and AWOL aggregations |
| `xf_phrase` | `title`, `phrase_text` | ticket status / priority / prefix reference cache |

#### NF Rosters add-on (milpacs)

| Table | Purpose | Key columns |
| --- | --- | --- |
| `xf_nf_rosters_user` | member milpac record | `relation_id`, `roster_id`, `user_id`, `username`, `position_id`, `secondary_position_ids`, `rank_id`, `bio`, `uniform_date`, `added_date`, `custom_fields` |
| `xf_nf_rosters_rank` | rank catalog | `rank_id`, `title`, `rank_image`, `display_order`, `extra_group_ids` |
| `xf_nf_rosters_position` | position catalog | `position_id`, `position_title`, `position_group_id`, `display_order`, `materialized_order`, `extra_group_ids`, `possible_secondary` |
| `xf_nf_rosters_position_group` | position/unit groups | `position_group_id`, `title`, `display_order` |
| `xf_nf_rosters_award` | award decoration catalog | `award_id`, `title`, `award_image`, `award_group_id`, `display_order`, `materialized_order` |
| `xf_nf_rosters_user_award` | member award records | `record_id`, `relation_id`, `award_id`, `from_user_id`, `details`, `award_date`, `citation_date` |
| `xf_nf_rosters_service_record` | member service records | `record_id`, `relation_id`, `details`, `record_date`, `citation_date`, `record_type_id` |
| `xf_nf_rosters_record_type` | record-type catalog | `record_type_id`, `title`, `display_order` |
| `xf_nf_rosters_field_value` | per-relation custom fields | `relation_id`, `field_id`, `field_value` |
| `xf_nf_rosters` | roster catalog | `roster_id`, `title`, `description`, `display_order`, `active` |

#### NF Tickets add-on

| Table | Purpose | Key columns |
| --- | --- | --- |
| `xf_nf_tickets_ticket` | ticket header | `ticket_id`, `ticket_ref`, `title`, `ticket_state`, `status_id`, `priority`, `prefix_id`, `assigned_user_id`, `ticket_category_id`, `starter_user_id`, reply/message timestamps, etc. |
| `xf_nf_tickets_message` | ticket posts | `message_id`, `ticket_id`, `user_id`, `username`, `message`, `message_date`, `message_state`, `position`, `attach_count`, `last_edit_date`, `edit_count` |
| `xf_nf_tickets_ticket_participant` | participant last-read | `ticket_id`, `user_id`, `last_read_date` |
| `xf_nf_tickets_ticket_field_value` | ticket custom fields | `ticket_id`, `field_id`, `field_value` |
| `xf_nf_tickets_category` | category tree | `ticket_category_id`, `title`, `description`, `parent_category_id`, `depth`, `lft`, `rgt`, `display_order`, `ticket_count` |
| `xf_phrase` | status/priority/prefix text | `title`, `phrase_text` |

#### Cav7 API Key add-on

| Table | Purpose | Key columns |
| --- | --- | --- |
| `xf_cav7_api_key` | key storage | `key_id`, `user_id`, `key_hash`, `key_prefix`, `is_active`, `created_date`, `last_used_date` |
| `xf_cav7_api_key_scope` | key-scope assignments | `key_id`, `scope_id` |
| `xf_cav7_api_key_scope_def` | scope definitions | `scope_id`, `scope_name` |

The integration-test schema is the most accurate reference for column types and indexes: <ref_file file="/home/bobbyswagg/research/DDI-API/testdb/schema.sql" />. Production indexes are documented separately in `testdb/indexes.sql`.

## 2. XenForo Dependency

### 2.1 Native XenForo tables used

The API depends on these stock tables:

* `xf_user` — canonical forum user identity
* `xf_user_group` — permission-group directory
* `xf_user_connected_account` — OAuth/connected account providers (Discord)
* `xf_post` — last-post aggregation for the AWOL and last-forum-post features
* `xf_phrase` — localized title resolution for the NF Tickets add-on

These are not owned by the API and would be present in any standard XenForo installation.

### 2.2 Custom 7CAV tables

The API relies on three add-on families, all prefixed in the database:

1. **NF Rosters** (`xf_nf_rosters_*`) — roster, rank, position, award, service-record, and per-member data.
2. **NF Tickets** (`xf_nf_tickets_*`) — support ticket system.
3. **Cav7 ApiKeyManager** (`xf_cav7_api_key*`) — the API's bearer token store and scope tables.

These tables are not native to XenForo. They define the data the API exposes.

### 2.3 API-specific tables

The only tables the API directly authenticates against are the `xf_cav7_api_key*` set. All other tables are upstream data sources the API maps to its wire types.

## 3. Data Models

The API maps these GORM models to the wire types in `types/`.

### 3.1 User / member model

`milpacs.Profile` maps to `xf_nf_rosters_user`:

| Field | Column(s) | Notes |
| --- | --- | --- |
| `RelationId` | `relation_id` | primary key (PK) for the milpac record |
| `RosterId` | `roster_id` | FK to `xf_nf_rosters` |
| `UserID` | `user_id` | FK to `xf_user.user_id` |
| `Username` | `username` | cached in rosters table |
| `PositionID` | `position_id` | FK to `xf_nf_rosters_position` |
| `SecondaryPositionIds` | `secondary_position_ids` | comma-separated list of `position_id` |
| `RankID` | `rank_id` | FK to `xf_nf_rosters_rank` |
| `Bio` | `bio` | free-form text |
| `UniformDate` | `uniform_date` | timestamp (uniform snapshot) |
| `AddedDate` | `added_date` | timestamp (record creation) |
| `CustomFields` | `custom_fields` | JSON blob with `realName`, `joinDate`, `promoDate`, `consoleGamertag`, `mos` |

Relationships:

```go
XfUser           XfUser           `gorm:"foreignKey:UserID;references:user_id"`
Primary          Position         `gorm:"foreignKey:PositionID;references:position_id"`
Rank             Rank             `gorm:"foreignKey:RankID;references:rank_id"`
Records          []Record         `gorm:"foreignKey:RelationID"`
AwardRecords     []AwardRecord    `gorm:"foreignKey:RelationID"`
ConnectedAccount []xenforo.ConnectedAccount `gorm:"foreignKey:UserID;references:user_id"`
```

<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/profiles.go" />

### 3.2 Rank model

`milpacs.Rank` → `xf_nf_rosters_rank`:

| Field | Column | Notes |
| --- | --- | --- |
| `RankId` | `rank_id` | PK |
| `Title` | `title` | full rank name |
| `RankImage` | `rank_image` | image cache-buster ID |
| `DisplayOrder` | `display_order` | sort order |
| `ExtraGroupIds` | `extra_group_ids` | additional forum group IDs |

The API exposes `Rank` and `RankExpanded` wire types. `types.RankShort` is derived from the `RankType` enum (`RANK_TYPE_...`).

<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/ranks.go" />  
<ref_file file="/home/bobbyswagg/research/DDI-API/types/ranks.go" />

### 3.3 Award model

`milpacs.Award` → `xf_nf_rosters_award` (catalog): `AwardId`, `Title`, `AwardImage`, `AwardGroupID`, `DisplayOrder`, `MaterializedOrder`.

`milpacs.AwardRecord` → `xf_nf_rosters_user_award` (member records):

| Field | Column |
| --- | --- |
| `RecordID` | `record_id` |
| `RelationID` | `relation_id` |
| `AwardID` | `award_id` |
| `FromUserID` | `from_user_id` |
| `Details` | `details` |
| `AwardDate` | `award_date` |
| `CitationDate` | `citation_date` |

<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/awards.go" />

### 3.4 Position model

`milpacs.Position` → `xf_nf_rosters_position`:

| Field | Column |
| --- | --- |
| `PositionId` | `position_id` |
| `PositionTitle` | `position_title` |
| `PositionGroupId` | `position_group_id` |
| `DisplayOrder` | `display_order` |
| `MaterializedOrder` | `materialized_order` |
| `ExtraGroupIds` | `extra_group_ids` |
| `PossibleSecondary` | `possible_secondary` |

`milpacs.PositionGroups` → `xf_nf_rosters_position_group`:

| Field | Column |
| --- | --- |
| `PositionGroupId` | `position_group_id` |
| `Title` | `title` |
| `DisplayOrder` | `display_order` |

<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/positions.go" />  
<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/position_groups.go" />

### 3.5 Roster model

`milpacs.Roster` → `xf_nf_rosters` (catalog):

| Field | Column |
| --- | --- |
| `RosterID` | `roster_id` |
| `Title` | `title` |
| `Description` | `description` |
| `DisplayOrder` | `display_order` |
| `Active` | `active` |

Members are assigned to a roster via `xf_nf_rosters_user.roster_id`.

<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/rosters.go" />

### 3.6 Record model

`milpacs.Record` → `xf_nf_rosters_service_record`:

| Field | Column |
| --- | --- |
| `RecordID` | `record_id` |
| `RelationID` | `relation_id` |
| `Details` | `details` |
| `RecordDate` | `record_date` |
| `CitationDate` | `citation_date` |
| `RecordTypeId` | `record_type_id` |

`milpacs.RecordType` → `xf_nf_rosters_record_type`: `RecordTypeId`, `Title`, `DisplayOrder`.

The `RecordType` enum in `types/recordtype.go` maps the numeric `record_type_id` to names.

<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/records.go" />  
<ref_file file="/home/bobbyswagg/research/DDI-API/types/recordtype.go" />

## 4. Authentication

### 4.1 API key storage

Keys live in the `xf_cav7_api_key*` tables:

* `xf_cav7_api_key` stores the key hash, key prefix, active flag, and usage timestamps.
* `xf_cav7_api_key_scope` links a key to one or more scopes.
* `xf_cav7_api_key_scope_def` defines the available scope names (`read`, `read:tickets`).

### 4.2 Permissions

Scopes are per-endpoint and checked by `rest.requireScope`. There are only two scopes today:

* `read` — access to milpacs, rosters, ranks, positions, AWOL, forum groups.
* `read:tickets` — access to the tickets surface.

<ref_snippet file="/home/bobbyswagg/research/DDI-API/rest/auth.go" lines="88-109" />

### 4.3 Token validation

`datastores/mysql.go` `ValidateApiKey` does the following:

1. Receives the raw token from the `Authorization: Bearer <token>` header.
2. Looks up `xf_cav7_api_key` where `key_hash = UNHEX(SHA2(?, 256))`, `is_active = 1`, and scope definition is active.
3. Joins `xf_cav7_api_key_scope` and `xf_cav7_api_key_scope_def` to collect allowed scope names.
4. Bumps `last_used_date` asynchronously on the matched `key_id`.

<ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="761-804" />

## 5. DDI Migration Options

### Option A: Keep XenForo/MariaDB backend initially

This means the API continues to read from a XenForo + NF Rosters + NF Tickets + ApiKeyManager database, just with DDI branding and (later) DDI data.

**Pros**

* Minimal code change in Phase 1/2/3: the GORM models, SQL joins, and reference-cache logic stay the same.
* `go test ./...` and the integration harness keep working.
* No data migration or ETL is required at this stage.
* The API can be pointed at a new DDI XenForo instance by changing `DB_*` and `FORUM_BASE_URL`.

**Cons**

* DDI remains coupled to XenForo and the NF Rosters add-on.
* The `xf_cav7_api_key*` table names still contain `Cav7` (a 7CAV brand). If that is unacceptable, the auth table set must be renamed, which means updating `datastores/mysql.go` and the test/schema fixtures.
* Rank, position, roster, and record values will still be 7CAV-specific until the upstream data is replaced.
* Hardcoded `https://7cav.us` image URLs remain in `milpacs/*.go` and golden responses.

### Option B: Create a standalone DDI database

This means designing a new schema (possibly still on MariaDB/MySQL) that the API reads directly, independent of XenForo.

**Pros**

* Full decoupling from XenForo and 7CAV-specific add-ons.
* Clean DDI branding and a schema tailored to DDI needs.
* Freedom to choose different auth, tickets, and roster systems.

**Cons**

* High risk and high effort: almost every GORM model and SQL query must be rewritten or remapped.
* The data layer has no migrations today, so DDI would need to create and own a new schema.
* Member identities, connected accounts (Discord), last-post aggregation, and ticket phrase lookups would need replacement implementations.
* Custom fields (`realName`, `mos`, etc.), rank catalog, position groups, and award catalogs would need redefinition.
* The golden contract corpus and integration tests would need a full reseed with DDI data.

### Recommendation

**Choose Option A for the immediate migration.**

The safest path is to keep the existing XenForo/MariaDB backend and simply point the API at a DDI-hosted XenForo instance with the same add-on set. The API has no DDL ownership and its GORM models are tightly coupled to the current table names. Option B should only be considered after the API is fully rebranded, the DDI data taxonomy (ranks, positions, rosters) is defined, and a separate migration of member/ticket data is planned.

If the `Cav7` auth table branding is a concern, the smallest incremental change under Option A is to create DDI-named copies of the `xf_cav7_api_key*` tables and update `datastores/mysql.go` queries and `testdb/schema.sql` fixtures — without touching the upstream data tables.