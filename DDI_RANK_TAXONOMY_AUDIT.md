# DDI Rank Taxonomy Audit

> Branch: `ddi-api-data-conversion`  
> Scope: rank IDs, wire output, image URLs, and migration strategy. No code changes.

This audit documents how ranks are represented in the code, on the wire, and in the database, and recommends the safest way to move to DDI rank terminology without changing rank IDs or image paths.

## 1. Current Rank IDs and Meanings

`types/ranks.go` defines a U.S. Army pay-grade enum. The numeric values are the `rank_id` foreign key in `xf_nf_rosters_rank` <ref_file file="/home/bobbyswagg/research/DDI-API/types/ranks.go" />:

| `rank_id` | `RankType` constant | Wire `RankShort` | Derivation |
| --- | --- | --- | --- |
| 0 | `RankTypeUnspecified` | `UNSPECIFIED` | `RANK_TYPE_UNSPECIFIED` minus `RANK_TYPE_` |
| 1 | `RankTypeGoa` | `GOA` | `RANK_TYPE_GOA` |
| 2 | `RankTypeGen` | `GEN` | `RANK_TYPE_GEN` |
| 3 | `RankTypeLtg` | `LTG` | `RANK_TYPE_LTG` |
| 4 | `RankTypeMg` | `MG` | `RANK_TYPE_MG` |
| 5 | `RankTypeBg` | `BG` | `RANK_TYPE_BG` |
| 6 | `RankTypeCol` | `COL` | `RANK_TYPE_COL` |
| 7 | `RankTypeLtc` | `LTC` | `RANK_TYPE_LTC` |
| 8 | `RankTypeMaj` | `MAJ` | `RANK_TYPE_MAJ` |
| 9 | `RankTypeCpt` | `CPT` | `RANK_TYPE_CPT` |
| 10 | `RankType1lt` | `1LT` | `RANK_TYPE_1LT` |
| 11 | `RankType2lt` | `2LT` | `RANK_TYPE_2LT` |
| 12 | `RankTypeCsm` | `CSM` | `RANK_TYPE_CSM` |
| 13 | `RankTypeSgm` | `SGM` | `RANK_TYPE_SGM` |
| 14 | `RankType1sg` | `1SG` | `RANK_TYPE_1SG` |
| 15 | `RankTypeMsg` | `MSG` | `RANK_TYPE_MSG` |
| 16 | `RankTypeSfc` | `SFC` | `RANK_TYPE_SFC` |
| 17 | `RankTypeSsg` | `SSG` | `RANK_TYPE_SSG` |
| 18 | `RankTypeSgt` | `SGT` | `RANK_TYPE_SGT` |
| 19 | `RankTypeCpl` | `CPL` | `RANK_TYPE_CPL` |
| 20 | `RankTypeSpc` | `SPC` | `RANK_TYPE_SPC` |
| 21 | `RankTypePfc` | `PFC` | `RANK_TYPE_PFC` |
| 22 | `RankTypePvt` | `PVT` | `RANK_TYPE_PVT` |
| 23 | `RankTypeRct` | `RCT` | `RANK_TYPE_RCT` |
| 24, 25 | (unassigned) | `24`, `25` | not in catalog; fallback to decimal string |
| 26 | `RankTypeCw5` | `CW5` | `RANK_TYPE_CW5` |
| 27 | `RankTypeCw4` | `CW4` | `RANK_TYPE_CW4` |
| 28 | `RankTypeCw3` | `CW3` | `RANK_TYPE_CW3` |
| 29 | `RankTypeCw2` | `CW2` | `RANK_TYPE_CW2` |
| 30 | `RankTypeWo1` | `WO1` | `RANK_TYPE_WO1` |
| 31 | `RankTypeAr` | `AR` | `RANK_TYPE_AR` |

The short name is produced by `RankShort()` which strips the `RANK_TYPE_` prefix from the enum name <ref_snippet file="/home/bobbyswagg/research/DDI-API/types/ranks.go" lines="119-124" />. This means the wire short name is fully determined by the `rank_id` value and the hardcoded `rankTypeNames` table.

The database table `xf_nf_rosters_rank` has the following fixture ranks in `testdb/fixtures.sql` <ref_snippet file="/home/bobbyswagg/research/DDI-API/testdb/fixtures.sql" lines="46-50" />:

| `rank_id` | `title` | `rank_image` |
| --- | --- | --- |
| 5 | `Sergeant Major` | 5 |
| 15 | `Sergeant` | 15 |
| 24 | `Specialist` | 24 |
| 27 | `Private` | 27 |

## 2. How Ranks Are Consumed by the API

### 2.1 Wire types

`types.Rank` and `types.S1UniformsRank` are the per-profile rank shapes:

```go
type Rank struct {
    RankShort    string `json:"rankShort"`
    RankFull     string `json:"rankFull"`
    RankImageUrl string `json:"rankImageUrl"`
    RankId       uint64 `json:"rankId,string"`
}
```

`types.RankExpanded` adds `RankDisplayOrder` for the rank catalog endpoint <ref_file file="/home/bobbyswagg/research/DDI-API/types/ranks.go" />.

### 2.2 Datastore mapping

`datastores/mysql.go` stamps `RankShort` onto every profile and rank catalog entry using the `RankType` enum:

```go
Rank: &types.Rank{
    RankId:       profile.RankID,
    RankShort:    types.RankType(profile.RankID).RankShort(),
    RankFull:     profile.Rank.Title,
    RankImageUrl: profile.Rank.ImageURL(),
}
```

<ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="136-141" />

`FindAllRanks` does the same for the catalog <ref_snippet file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" lines="488-495" />.

### 2.3 Rank catalog endpoint

`GET /api/v1/milpacs/ranks` returns `RanksResponse` — a list of `RankExpanded` ordered by `display_order` <ref_file file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" />.

### 2.4 Per-profile rank emission

`Rank` is embedded in:

* `Profile`
* `LiteProfile`
* `S1UniformsProfile`
* `RankExpanded`

It is also included in `Position` search results.

### 2.5 Test coverage

`types/ranktype_test.go` pins the `RANK_TYPE_` strings and the `RankShort` derivation <ref_file file="/home/bobbyswagg/research/DDI-API/types/ranktype_test.go" />. `types/ranks_test.go` tests the `RankExpanded` JSON shape. The contract goldens and harness tests assert the actual `RankImageUrl` values.

## 3. Relationship Between Rank IDs and Images

`milpacs/ranks.go` defines the `Rank` GORM model and its `ImageURL()` method <ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/ranks.go" />:

```go
func (rank *Rank) ImageURL() string {
    imageGroup := math.Floor(float64(rank.RankId / 1000))
    return fmt.Sprintf("%s/data/roster_ranks/%d/%d.jpg?%d",
        assetBaseURL(), int(imageGroup), rank.RankId, rank.RankImage)
}
```

The URL has four parts that are all driven by the `rank_id` and the `rank_image` column:

* `assetBaseURL()` — from `ASSET_BASE_URL` env (Phase 3.1)
* `imageGroup` — `rank_id / 1000` (0 for every current 7CAV rank)
* `rank_id` — the final path segment
* `rank_image` — the query-string cache buster

This means the image asset must live at `{assetBaseURL}/data/roster_ranks/{group}/{rank_id}.jpg?{rank_image}`. The path is rigid; only the `ASSET_BASE_URL` root is configurable. The `rank_image` value can be changed per row to bust caches without renaming files.

## 4. Which Parts Are Safe to Rename

### Safe to change (no code or schema change)

* `xf_nf_rosters_rank.title` — this is `RankFull` on the wire. DDI can put its own full rank names here.
* `xf_nf_rosters_rank.rank_image` — cache-buster value; changing it changes the query string but not the file path.
* The actual image files on the CDN, as long as they stay under the same `rank_id` and `group` paths.

### Not safe to change without code work

* `rank_id` values — they are the canonical foreign key in `xf_nf_rosters_user.rank_id`, the `types.RankType` enum, and the image path. Changing them requires updating rows, `RankType`, and the image directories.
* `types.RankType` / `rankTypeNames` — the `RankShort` derivation is hardcoded from `RANK_TYPE_` names. Replacing this requires a new mapping function.
* The image path template `data/roster_ranks/{group}/{rank_id}.jpg` — changing it requires modifying `milpacs/ranks.go` and all golden expectations.

### Safe to change with a config layer

The cleanest DDI-friendly path is to **replace the `RankShort()` derivation with a `RANK_SHORT_NAMES` env map**. This keeps `rank_id` and image paths stable while allowing DDI to control `rankShort` on the wire.

## 5. Recommended DDI Rank Migration Strategy

### 5.1 Keep rank IDs and image paths

The numeric `rank_id` should be treated as opaque. DDI should not renumber; instead, map each `rank_id` to a DDI rank title and short name.

### 5.2 Add a `RANK_SHORT_NAMES` configuration

Introduce a `RANK_SHORT_NAMES` env var with a JSON map from `rank_id` to short name:

```json
{
  "1": "GOA",
  "2": "GEN",
  ...
}
```

Update `datastores/mysql.go` to use `config.RankShort(rankId)` instead of `types.RankType(rankId).RankShort()`. Default to the current `RANK_TYPE_` derivation for unconfigured IDs so existing tests pass.

### 5.3 Add a `RANK_FULL_NAMES` configuration (optional)

If DDI wants `RankFull` to be a configured string independent of `xf_nf_rosters_rank.title`, add a similar map. Otherwise, DDI can simply update the `title` column in `xf_nf_rosters_rank`. Both are safe.

### 5.4 Add `RANK_IMAGE_CACHE` support (optional)

The `rank_image` column is already the cache buster. DDI can change it as assets are uploaded. If DDI wants per-rank image cache values driven by env, a `RANK_IMAGES` map could override `rank_image`, but this is usually unnecessary.

### 5.5 Keep `RankImageUrl` path structure

Leave the `milpacs/ranks.go` `ImageURL()` template unchanged. Upload DDI rank images to the same `{group}/{rank_id}.jpg` locations under the new `ASSET_BASE_URL`.

### 5.6 Update fixture and golden data

* `testdb/fixtures.sql` `xf_nf_rosters_rank` titles and `rank_image` values
* `contract/goldens/milpacs/ranks.golden.json` for rank catalog
* `types/ranktype_test.go` only if the `RANK_TYPE_` derivation is replaced by a config map

## 6. Files to Touch in a Future Implementation Phase

If adding `RANK_SHORT_NAMES`:

* `internal/config/orgmap.go` (or new `ranks.go`) to load the map
* `datastores/mysql.go` `FindAllRanks` and profile mappers
* `types/ranks.go` `RankShort()` or its replacement
* `types/ranktype_test.go` if the default derivation changes

If only changing titles:

* `testdb/fixtures.sql`
* `xf_nf_rosters_rank` production data

## 7. What Not to Change

Per the constraints:

* `xf_nf_rosters_rank` table schema
* `rank_id` values
* Image path template `data/roster_ranks/{group}/{rank_id}.jpg`
* `Rank` JSON field names (`rankShort`, `rankFull`, `rankImageUrl`, `rankId`)
* `RankExpanded` JSON field names