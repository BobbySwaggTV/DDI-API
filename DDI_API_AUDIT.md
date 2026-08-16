# DDI API Audit

> Branch: `ddi-api-audit`  
> Repo: `7Cav/api` (to become Dagger Defense Industries / DDI API)  
> Date: 2026-08-16

This is a read-only audit of the repository before any migration work. No code was changed.

## 1. Executive Summary

The service is a Go HTTP/JSON API built on the standard-library `net/http` stack. It is a read-only layer over a XenForo MySQL/MariaDB database augmented by the `NF Rosters` and `NF Tickets` add-ons. Authentication is Bearer-token based against a `Cav7` API-key add-on. The public surface, wire types, and documentation are maintained as a hand-written OpenAPI 3.1 spec and a frozen golden corpus.

The migration to DDI will mostly be a re-branding and data-layer translation exercise: the routing, middleware, and request/response plumbing are generic, while the rank/position/roster taxonomies, 7Cav asset URLs, `Cav7` auth tables, and 7Cav-specific position-group labels are tightly coupled to the source organization.

## 2. Repository Structure

### 2.1 Backend framework

* Language / runtime: Go 1.25.10 <ref_file file="/home/bobbyswagg/research/DDI-API/go.mod" />
* Module path: `github.com/7cav/api` <ref_file file="/home/bobbyswagg/research/DDI-API/go.mod" />
* HTTP stack: stdlib `net/http` with Go 1.22+ pattern routing <ref_file file="/home/bobbyswagg/research/DDI-API/rest/rest.go" />
* CLI: `cobra` + `viper` for env-only configuration <ref_file file="/home/bobbyswagg/research/DDI-API/cmd/root.go" />
* ORM: GORM 1.31.2 with `gorm.io/driver/mysql` <ref_file file="/home/bobbyswagg/research/DDI-API/go.mod" />
* Observability: Prometheus metrics, Sentry error capture (env-gated), gzip compression <ref_file file="/home/bobbyswagg/research/DDI-API/rest/rest.go" />
* Documentation: Scalar UI serving a hand-owned OpenAPI 3.1 spec <ref_file file="/home/bobbyswagg/research/DDI-API/rest/docs.go" />

### 2.2 Entry points

* `main.go` calls `cmd.Execute()` <ref_file file="/home/bobbyswagg/research/DDI-API/main.go" />
* `cmd/root.go` sets up `cobra` and `viper.AutomaticEnv()` <ref_file file="/home/bobbyswagg/research/DDI-API/cmd/root.go" />
* `cmd/serve.go` creates and starts `servers.MicroServer` <ref_file file="/home/bobbyswagg/research/DDI-API/cmd/serve.go" />
* `servers/server.go` builds the public listener on `:11000` and internal metrics listener on `:9090` <ref_file file="/home/bobbyswagg/research/DDI-API/servers/server.go" />

### 2.3 Package layout

| Package | Purpose |
| --- | --- |
| `cmd/` | CLI root and `serve` command |
| `contract/` | Golden-corpus contract tests and test batteries |
| `datastores/` | `Datastore` interface and MySQL/GORM implementation |
| `docs/`, `openapi/` | OpenAPI spec and ADRs |
| `internal/` | Shared OpenAPI contract-test helpers |
| `milpacs/` | GORM models for the NF Rosters add-on |
| `referencecache/` | In-memory reference cache for ticket categories/status/priority/prefix |
| `rest/` | Handlers, middleware, route registration, query binding |
| `servers/` | `MicroServer`, DB connection setup, connection-pool config |
| `testdb/` | Dockerized MariaDB harness, schema, fixtures, indexes |
| `types/` | Hand-written JSON wire types |
| `xenforo/` | GORM models for core XenForo and ticket tables |

### 2.4 Configuration system

All configuration is env-var based via `viper.AutomaticEnv()` <ref_file file="/home/bobbyswagg/research/DDI-API/cmd/root.go" />.

Required:

* `DB_USERNAME`
* `DB_PASSWORD`
* `DB_HOST`
* `DB_PORT`

Optional / operational:

* `DB_MAX_OPEN_CONNS`
* `DB_MAX_IDLE_CONNS`
* `DB_CONN_MAX_LIFETIME` <ref_file file="/home/bobbyswagg/research/DDI-API/servers/pool.go" />
* `TRUSTED_PROXIES` <ref_file file="/home/bobbyswagg/research/DDI-API/rest/clientip.go" />
* `SENTRY_DSN`
* `SENTRY_DEBUG`
* `FORUM_BASE_URL` <ref_file file="/home/bobbyswagg/research/DDI-API/datastores/tickets.go" />
* `REFERENCE_CACHE_REFRESH_INTERVAL` <ref_file file="/home/bobbyswagg/research/DDI-API/servers/server.go" />

The database name is hardcoded as `xenforo` in the DSN <ref_snippet file="/home/bobbyswagg/research/DDI-API/servers/server.go" lines="99" />.

## 3. Database Architecture

### 3.1 Database engine

* MariaDB / MySQL accessed through GORM.
* Database name: `xenforo`.
* The API issues no DDL. The production schema is owned by the upstream XenForo forum and its add-ons.
* Recommended indexes for performance are documented in `testdb/indexes.sql` and applied manually by the DB admin <ref_file file="/home/bobbyswagg/research/DDI-API/CONTEXT.md" />.

### 3.2 Schema overview

The integration-test schema in `testdb/schema.sql` is the best reference for the tables the API reads <ref_file file="/home/bobbyswagg/research/DDI-API/testdb/schema.sql" />.

#### Core XenForo

| Table | Purpose |
| --- | --- |
| `xf_user` | Forum users (`user_id`, `username`) |
| `xf_user_group` | Forum permission groups (`user_group_id`, `title`) |
| `xf_user_connected_account` | Connected accounts (Discord, etc.) by `provider` / `provider_key` |
| `xf_post` | Forum posts (used for last-post and AWOL aggregations) |
| `xf_phrase` | Localized phrases (used to resolve ticket status/priority/prefix) |

#### NF Rosters add-on (milpacs)

| Table | Purpose |
| --- | --- |
| `xf_nf_rosters_user` | Member milpac record — one row per member/roster assignment <ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/profiles.go" /> |
| `xf_nf_rosters_rank` | Rank catalog (`rank_id`, `title`, `rank_image`, `display_order`) <ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/ranks.go" /> |
| `xf_nf_rosters_position` | Position catalog (`position_id`, `position_title`, `position_group_id`, `possible_secondary`) <ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/positions.go" /> |
| `xf_nf_rosters_position_group` | Position groups (units / billet groups) <ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/position_groups.go" /> |
| `xf_nf_rosters_award` | Award decoration catalog <ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/awards.go" /> |
| `xf_nf_rosters_user_award` | Member award records <ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/awards.go" /> |
| `xf_nf_rosters_service_record` | Member service-record entries <ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/records.go" /> |
| `xf_nf_rosters_record_type` | Service-record type catalog <ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/records.go" /> |
| `xf_nf_rosters_field_value` | Custom fields per milpac relation (e.g. `realName`, `mos`, `consoleGamertag`) <ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/field_value.go" /> |
| `xf_nf_rosters` | Roster catalog itself (roster types) <ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/rosters.go" /> |

#### NF Tickets add-on

| Table | Purpose |
| --- | --- |
| `xf_nf_tickets_ticket` | Ticket header data |
| `xf_nf_tickets_message` | Ticket messages |
| `xf_nf_tickets_ticket_participant` | Participants and last-read dates |
| `xf_nf_tickets_ticket_field_value` | Custom fields per ticket |
| `xf_nf_tickets_category` | Ticket category tree |

#### Cav7 API Key add-on

| Table | Purpose |
| --- | --- |
| `xf_cav7_api_key` | API keys (SHA256 hash) |
| `xf_cav7_api_key_scope` | Key-to-scope linkage |
| `xf_cav7_api_key_scope_def` | Scope definitions (`read`, `read:tickets`) <ref_file file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" /> |

### 3.3 Key table columns

`xf_nf_rosters_user` (member table):

```
relation_id              (PK)
roster_id
user_id
username
position_id
secondary_position_ids   (comma-separated)
rank_id
bio
uniform_date
added_date
custom_fields            (JSON blob)
```

<ref_file file="/home/bobbyswagg/research/DDI-API/testdb/schema.sql" />

Custom fields are stored as a JSON blob with known keys: `realName`, `joinDate`, `promoDate`, `consoleGamertag`, `mos` <ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/profiles.go" />.

## 4. API Endpoints

All API routes are mounted under `/api/v1` in `rest.routes()` <ref_file file="/home/bobbyswagg/research/DDI-API/rest/rest.go" />. Every endpoint requires `Authorization: Bearer <token>`. Scopes are `read` (milpacs / forum surface) or `read:tickets` (tickets surface). Error responses are gRPC-style JSON: `{"code":N,"message":"...","details":[]}` <ref_file file="/home/bobbyswagg/research/DDI-API/rest/write.go" />.

### 4.1 Milpacs / roster surface (`read` scope, `Cache-Control: max-age=600`)

| Method | Path | Purpose | Parameters | Response |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/milpacs/ranks` | Rank catalog | none | `RanksResponse` — `ranks[]` of `RankExpanded` |
| GET | `/api/v1/milpacs/position/groups` | Position-group tree | none | `PositionGroupsResponse` — `groups[]` of `PositionGroup` |
| GET | `/api/v1/milpacs/awol` | AWOL report | none | `AwolResponse` — `awols[]` of `Awol` |
| GET | `/api/v1/milpacs/position/search/{positionQuery...}` | Search lite profiles by position title substring (also `/position/search` for empty query) | path: `positionQuery` | `LiteRoster` — `profiles{}` keyed by relation id |
| GET | `/api/v1/milpacs/profile/id/{userId}` | Full profile by milpac relation id | path: `userId` (64-bit, parsed base-0); query: `username` (legacy override) | `Profile` |
| GET | `/api/v1/milpacs/profile/username/{username}` | Full profile by username | path: `username`; query: `userId` (ignored legacy) | `Profile` |
| GET | `/api/v1/milpac/discord/{discordId}` | Full profile by Discord id | path: `discordId` | `Profile` |
| GET | `/api/v1/milpac/gamertag/{gamertag}` | Full profile by console gamertag | path: `gamertag` | `Profile` |
| GET | `/api/v1/roster/{roster}` | Full roster | path: `roster` (name or number, e.g. `ROSTER_TYPE_COMBAT` or `1`) | `Roster` — `profiles{}` of `Profile` |
| GET | `/api/v1/roster/{roster}/lite` | Lite roster (no records/awards) | path: `roster` | `LiteRoster` — `profiles{}` of `LiteProfile` |
| GET | `/api/v1/s1/uniforms/{roster}` | Uniform-builder roster view | path: `roster` | `S1UniformsRoster` — `profiles{}` of `S1UniformsProfile` |

<ref_file file="/home/bobbyswagg/research/DDI-API/openapi/openapi.yaml" />

### 4.2 Forum reference surface (`read` scope, `Cache-Control: max-age=600`)

| Method | Path | Purpose | Parameters | Response |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/forum/groups` | Forum permission-group directory | none | `[]ForumGroup` (array of `{groupId, groupName}`) |

### 4.3 Tickets surface (`read:tickets` scope, `Cache-Control: max-age=0`)

| Method | Path | Purpose | Parameters | Response |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/tickets` | Cursor-paginated ticket list | query: `categoryId[]`, `excludeSubcategories`, `ticketState[]`, `statusId[]`, `prefixId[]`, `assignedUserId[]`, `starterUserId[]`, `modifiedSince`, `includeHidden`, `perPage`, `afterCursor` | `ListTicketsResponse` |
| GET | `/api/v1/tickets/categories` | Ticket category tree | none | `ListCategoriesResponse` |
| GET | `/api/v1/tickets/{ticketId}` | Ticket + first 10 messages | path: `ticketId` (uint32) | `GetTicketResponse` |
| GET | `/api/v1/tickets/ref/{ticketRef}` | Ticket by alphanumeric ref | path: `ticketRef` | `GetTicketResponse` |
| GET | `/api/v1/tickets/{ticketId}/messages` | Paginated messages | path: `ticketId`; query: `perPage`, `afterCursor`, `includeHidden` | `ListTicketMessagesResponse` |
| GET | `/api/v1/tickets/ref/messages` | Frozen 400 parity shim | none | `400` error |

<ref_file file="/home/bobbyswagg/research/DDI-API/rest/tickets.go" />

### 4.4 Public docs surface (no API key)

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/openapi.yaml` | The OpenAPI 3.1 spec |
| GET | `/` and other non-`/api` paths | Scalar documentation UI |

### 4.5 Wire types quick reference

* `Profile` — full member: `user`, `rank`, `realName`, `uniformUrl`, `roster`, `primary`, `secondaries[]`, `records[]`, `awards[]`, `joinDate`, `promotionDate`, `discordId`, `lastForumPostDate`, `mos`, `consoleGamertag` <ref_file file="/home/bobbyswagg/research/DDI-API/types/profiles.go" />
* `LiteProfile` — same minus records/awards plus `awardDate`, `recordDate` <ref_file file="/home/bobbyswagg/research/DDI-API/types/profiles.go" />
* `S1UniformsProfile` — uniform-audit view: `user`, `rank`, `realName`, `uniformUrl`, `uniformDate`, `uniformUpdateTriggerDate`, `roster`, `primaryPositionTitle`, `secondaries[]`, `joinDate`, `promotionDate`, `areaOfResponsibility` <ref_file file="/home/bobbyswagg/research/DDI-API/types/profiles.go" />
* `RosterType` enum values: `ROSTER_TYPE_UNSPECIFIED`, `ROSTER_TYPE_COMBAT` (1), `ROSTER_TYPE_RESERVE` (2), `ROSTER_TYPE_ELOA` (3), `ROSTER_TYPE_WALL_OF_HONOR` (4), `ROSTER_TYPE_ARLINGTON` (5), `ROSTER_TYPE_PAST_MEMBERS` (6) <ref_file file="/home/bobbyswagg/research/DDI-API/types/roster.go" />
* `RecordType` enum values: `PROMOTION`, `OPERATION`, `TRANSFER`, `DISCIPLINARY`, `DISCHARGE`, `ASSIGNMENT`, `NAME_CHANGE`, `ELOA`, `GRADUATION` <ref_file file="/home/bobbyswagg/research/DDI-API/types/recordtype.go" />
* `RankType` enum: U.S. Army-style ranks from `GOA` to `AR` plus warrant ranks <ref_file file="/home/bobbyswagg/research/DDI-API/types/ranks.go" />

## 5. CavApps / Uniform Builder Compatibility

The source repository does not contain the string "CavApps" or "uniform builder". The relevant consumer is the *S1 uniforms tool*, which is served by `GET /api/v1/s1/uniforms/{roster}` <ref_file file="/home/bobbyswagg/research/DDI-API/rest/rosters.go" />.

### 5.1 Uniform-builder JSON structure

The endpoint returns an `S1UniformsRoster`:

```json
{
  "profiles": {
    "1": {
      "user": { "userId": "1", "username": "..." },
      "rank": { "rankShort": "MG", "rankFull": "Major General", "rankImageUrl": "..." },
      "realName": "...",
      "uniformUrl": "...",
      "uniformDate": "...",
      "uniformUpdateTriggerDate": "...",
      "roster": "ROSTER_TYPE_COMBAT",
      "primaryPositionTitle": "...",
      "secondaries": [{ "positionTitle": "..." }],
      "joinDate": "...",
      "promotionDate": "...",
      "areaOfResponsibility": "..."
    }
  }
}
```

<ref_file file="/home/bobbyswagg/research/DDI-API/types/profiles.go" />

### 5.2 Which endpoints provide each data category

| Category | Endpoints providing it | Notes |
| --- | --- | --- |
| **user** | `GET /api/v1/milpacs/profile/id/{userId}`<br>`GET /api/v1/milpacs/profile/username/{username}`<br>`GET /api/v1/milpac/discord/{discordId}`<br>`GET /api/v1/milpac/gamertag/{gamertag}`<br>`GET /api/v1/roster/{roster}`<br>`GET /api/v1/roster/{roster}/lite`<br>`GET /api/v1/s1/uniforms/{roster}`<br>`GET /api/v1/milpacs/position/search/{positionQuery}` | Embedded `User` object (`userId` as string, `username`) |
| **rank** | `GET /api/v1/milpacs/ranks` (full catalog)<br>All profile/roster endpoints above (per-member `Rank`) | `RankExpanded` adds `rankDisplayOrder` |
| **MOS** | `GET /api/v1/milpacs/profile/id/{userId}`<br>`GET /api/v1/milpacs/profile/username/{username}`<br>`GET /api/v1/milpac/discord/{discordId}`<br>`GET /api/v1/milpac/gamertag/{gamertag}`<br>`GET /api/v1/roster/{roster}`<br>`GET /api/v1/roster/{roster}/lite`<br>`GET /api/v1/milpacs/position/search/{positionQuery}` | `mos` field from custom fields; **not present** in `S1UniformsProfile` |
| **awards** | `GET /api/v1/milpacs/profile/...`<br>`GET /api/v1/roster/{roster}` | `awards[]` on full `Profile`; absent from `LiteProfile` and `S1UniformsProfile` |
| **roster information** | `GET /api/v1/roster/{roster}` (full)<br>`GET /api/v1/roster/{roster}/lite`<br>`GET /api/v1/s1/uniforms/{roster}`<br>`GET /api/v1/milpacs/position/search/{positionQuery}` (lite subset)<br>`GET /api/v1/milpacs/awol` (absence report) | Main roster endpoints return a `profiles` map keyed by milpac relation id |

For a uniform-builder that only needs rank/position/uniform metadata, `GET /api/v1/s1/uniforms/{roster}` is the dedicated shape. If it also needs MOS or awards, it must fall back to `GET /api/v1/roster/{roster}` (full) or `GET /api/v1/milpacs/profile/id/{userId}`.

## 6. 7Cav Dependencies

### 6.1 Module and import paths

* `go.mod` declares `module github.com/7cav/api`.
* Every Go file imports `github.com/7cav/api/...`.
* GNU GPL copyright headers in every source file refer to `7Cav.us` and `7Cav-API`.

### 6.2 Hardcoded URLs and branding

| URL / brand | Location | Impact |
| --- | --- | --- |
| `https://api.7cav.us` | `openapi/openapi.yaml` servers list <ref_snippet file="/home/bobbyswagg/research/DDI-API/openapi/openapi.yaml" lines="33" /> | API base URL in docs |
| `https://7cav.us/account/api-key` | `openapi/openapi.yaml` description <ref_snippet file="/home/bobbyswagg/research/DDI-API/openapi/openapi.yaml" lines="25" /> | Auth instructions |
| `https://github.com/7cav/api` | `openapi/openapi.yaml` external docs <ref_snippet file="/home/bobbyswagg/research/DDI-API/openapi/openapi.yaml" lines="31" /> | Repo link |
| `https://7cav.us/data/roster_ranks/%d/%d.jpg?%d` | `milpacs/ranks.go` <ref_snippet file="/home/bobbyswagg/research/DDI-API/milpacs/ranks.go" lines="38-41" /> | Rank image URLs served to clients |
| `https://7cav.us/data/roster_uniforms/%d/%d.jpg` | `milpacs/profiles.go` <ref_snippet file="/home/bobbyswagg/research/DDI-API/milpacs/profiles.go" lines="56-59" /> | Uniform image URLs |
| `https://7cav.us/data/roster_awards/%d/%d.jpg?%d` | `milpacs/awards.go` <ref_snippet file="/home/bobbyswagg/research/DDI-API/milpacs/awards.go" lines="39-42" /> | Award image URLs |
| `https://7cav.us/tickets/{ref}/` | `datastores/tickets.go` when `FORUM_BASE_URL=https://7cav.us` <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/tickets.go" lines="265-267" /> | Ticket `forumUrl` field |
| `7cav/api:latest`, `7cav-api` container | `docker-compose.yaml` <ref_snippet file="/home/bobbyswagg/research/DDI-API/docker-compose.yaml" lines="15-16" /> | Docker image / container name |
| `logo.png` | repository root | 7Cav logo used by README |

### 6.3 Authentication dependency

* API key storage: `xf_cav7_api_key`, `xf_cav7_api_key_scope`, `xf_cav7_api_key_scope_def` <ref_file file="/home/bobbyswagg/research/DDI-API/testdb/schema.sql" />.
* Token validation: `UNHEX(SHA2(?, 256))` against `key_hash` <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="767-774" />.
* Test keys use the `cav7_` prefix (`cav7_readkey`, `cav7_ticketskey`, etc.) throughout the test suite.
* Scopes are `read` and `read:tickets`.

### 6.4 Discord / connected-account integration

* Discord lookup filters `xf_user_connected_account.provider = 'nfDiscord'` <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="109-110" />.
* Discord id extraction likewise checks `provider == "nfDiscord"` <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="162-169" />.

### 6.5 Organization-specific business logic

* `getPositionGroup()` in `datastores/mysql.go` maps 7Cav position-group titles to area-of-responsibility strings:

```go
"Regimental HQ", "Support Attachment" -> "HHQ"
"New Recruits"                        -> "RTC"
"7th Cavalry Reservists"              -> "Reserve"
"Extended Leave Of Absence"           -> "ELOA"
Any title containing "Command"        -> "HHQ"
```

<ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="457-475" />

* Roster, rank, and record enumerations are U.S. Army / 7Cav specific (e.g. `ROSTER_TYPE_COMBAT`, `RANK_TYPE_CSM`, `RECORD_TYPE_PROMOTION`).
* Test fixtures and golden corpus include 7Cav sample data (`Regimental HQ`, `Alpha Company`, `7th Cavalry Reservists`, rank images from `7cav.us`, etc.).

### 6.6 Summary of 7Cav-tagged files

A broad search for `7cav`, `7Cav`, `7th Cavalry`, or `7Cavalry` matches 100+ files, including:

* All `.go` source and test files (copyright headers and package imports)
* `go.mod`
* `README.md`
* `openapi/openapi.yaml` and assets
* `contract/goldens/*` (fixtures / sample responses)
* `testdb/schema.sql`, `testdb/fixtures.sql`
* `Dockerfile`, `docker-compose.yaml`
* `milpacs/*.go` (rank/uniform/award image URL builders)

## 7. DDI Conversion Recommendations

### 7.1 What can remain unchanged

* The stdlib `net/http` + `http.ServeMux` routing stack and middleware chain.
* The `Datastore` interface and GORM-backed MySQL implementation pattern.
* The hand-written `types` package and wire-convention rules (lowerCamelCase, emit-everything, `,string` for `int64`, etc.).
* The golden-corpus contract flow (`contract/`, `openapi/openapi.yaml`, `contract/spec_test.go`).
* The cache-control / auth / gzip / metrics / Sentry middleware (after re-branding only the names).
* The connection-pool hardening and env configuration plumbing.
* The `referencecache` pattern for ticket reference data.

### 7.2 What needs configuration changes

* Go module path: `github.com/7cav/api` -> DDI module path.
* All `github.com/7cav/api/...` imports updated.
* `openapi/openapi.yaml`: title, description, `info.version`, `servers` URL, auth manager link, GitHub repo link.
* `README.md`, `docker-compose.yaml`, `Dockerfile` labels/image names.
* `FORUM_BASE_URL` env var to point at the DDI forum.
* `SENTRY_DSN` (optional) and `TRUSTED_PROXIES` (required for prod reverse-proxy).
* Static assets: swap `logo.png` for DDI branding.

### 7.3 What needs rewriting

| Component | Why | Priority |
| --- | --- | --- |
| **Image URL builders** | `milpacs/ranks.go`, `milpacs/profiles.go`, `milpacs/awards.go` hardcode `https://7cav.us/data/...` | High — served in every profile response |
| **Position-group AOR mapping** | `datastores/mysql.go` `getPositionGroup` hardcodes 7Cav unit titles | High — affects `S1UniformsProfile` |
| **Rank / roster / record enums** | `types/ranks.go`, `types/roster.go`, `types/recordtype.go` encode 7Cav taxonomy | High if DDI uses different ranks/rosters |
| **Auth datastore** | `datastores/mysql.go` `ValidateApiKey` uses `xf_cav7_api_key*` tables and `cav7_` key format | High — every API call |
| **Connected-account provider** | Discord/gamertag lookups assume `nfDiscord` and `consoleGamertag` field id | Medium — only if DDI keeps Discord lookups |
| **Ticket phrase families** | `datastores/tickets.go` reads `nf_tickets_status.`, `nf_tickets_priority.`, `nf_tickets_ticket_prefix.` from `xf_phrase` | Medium — only if DDI keeps NF Tickets |
| **Custom fields** | `milpacs/profiles.go` `CustomFields` struct assumes `realName`, `joinDate`, `promoDate`, `consoleGamertag`, `mos` | Medium — map to DDI fields |
| **Test fixtures / goldens** | `testdb/fixtures.sql`, `contract/goldens/*` contain 7Cav data and URLs | Medium — must be regenerated for DDI |
| **Copyright headers** | Every `.go` file contains 7Cav GPL header | Low — legal/re-branding task |

### 7.4 Suggested migration order

1. **Fork and rename** the Go module and all internal imports.
2. **Re-brand** `README.md`, `openapi/openapi.yaml`, `docker-compose.yaml`, `Dockerfile`, and `logo.png`.
3. **Externalize asset URLs**: make the rank/uniform/award image base URL a configurable env var (`ASSET_BASE_URL` or similar) and remove the `7cav.us` builders.
4. **Decide the DDI database schema**: keep the NF Rosters/XenForo table names or map DDI tables to the existing GORM models.
5. **Update domain enumerations**: replace or extend `RosterType`, `RankType`, `RecordType` for DDI values.
6. **Update `getPositionGroup`** with DDI unit/position-group naming.
7. **Replace the `Cav7` auth implementation**: point `ValidateApiKey` at a DDI auth store and remove `cav7_` key assumptions.
8. **Update connected-account providers and custom-field mapping** for DDI identity sources.
9. **Rewrite test fixtures and the golden corpus** with DDI sample data; run `go test ./...` and `make test-integration`.
10. **Smoke test** the public endpoints against a DDI-shaped database.

## 8. Notes and Caveats

* The service has no migration / schema-ownership tooling; it assumes a pre-existing XenForo + add-on database.
* The API is read-only, but `ValidateApiKey` writes `last_used_date` asynchronously.
* The route `GET /api/v1/milpac/discord/{discordId}` uses the historical singular `milpac` path (not `milpacs`); this is frozen by the golden corpus and should not be "fixed" without updating goldens.
* `RosterType` path parameters accept either the enum name (`ROSTER_TYPE_COMBAT`) or the underlying integer `1` (Go base-0 parsing), which means hex/octal/binary literals are also accepted. Changing this requires updating the golden corpus.
* The uniform-builder specific view deliberately omits `mos` and `awards`; if the DDI uniform tool needs those fields, use the full `Profile` or `LiteProfile` surfaces.