# DDI Rank Asset Requirements

> Phase 5.5A: preparation for DDI-owned rank artwork.  
> No existing artwork was replaced, no rank IDs were renamed, no API contracts were changed, and no uniform-rendering code was modified.

## 1. Required Rank IDs

The API, the BFF, and the uniform builder use the following `rank_id` values. The gaps at 24 and 25 are intentionally unassigned and must be left empty.

| rank_id | API `rankShort` | BFF `RANK_SHORT` | Uniform `rankGrade` | Grade meaning |
| --- | --- | --- | --- | --- |
| 1 | GOA | GA | O11 | General of the Army |
| 2 | GEN | GEN | O10 | General |
| 3 | LTG | LTG | O9 | Lieutenant General |
| 4 | MG | MG | O8 | Major General |
| 5 | BG | BG | O7 | Brigadier General |
| 6 | COL | COL | O6 | Colonel |
| 7 | LTC | LTC | O5 | Lieutenant Colonel |
| 8 | MAJ | MAJ | O4 | Major |
| 9 | CPT | CPT | O3 | Captain |
| 10 | 1LT | 1LT | O2 | First Lieutenant |
| 11 | 2LT | 2LT | O1 | Second Lieutenant |
| 12 | CSM | CSM | E11 | Command Sergeant Major |
| 13 | SGM | SGM | E10 | Sergeant Major |
| 14 | 1SG | 1SG | E9 | First Sergeant |
| 15 | MSG | MSG | E8 | Master Sergeant |
| 16 | SFC | SFC | E7 | Sergeant First Class |
| 17 | SSG | SSG | E6 | Staff Sergeant |
| 18 | SGT | SGT | E5 | Sergeant |
| 19 | CPL | CPL | E4 | Corporal |
| 20 | SPC | SPC | E3 | Specialist |
| 21 | PFC | PFC | E2 | Private First Class |
| 22 | PVT | PVT | E1 | Private |
| 23 | RCT | RCT | E0 | Recruit |
| 24 | — | — | — | **unassigned, must remain empty** |
| 25 | — | — | — | **unassigned, must remain empty** |
| 26 | CW5 | CW5 | W5 | Chief Warrant Officer 5 |
| 27 | CW4 | CW4 | W4 | Chief Warrant Officer 4 |
| 28 | CW3 | CW3 | W3 | Chief Warrant Officer 3 |
| 29 | CW2 | CW2 | W2 | Chief Warrant Officer 2 |
| 30 | WO1 | WO1 | W1 | Warrant Officer 1 |
| 31 | AR | AR | E0 | Aviation Recruit (or similar) |
| 32 | TST | TST | E0 | Test rank (BFF only, may not have art) |

<ref_file file="/home/bobbyswagg/research/DDI-API/internal/config/orgmap.go" />
<ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/GetUserInfo.jsx" />

## 2. Required API Image Paths

The DDI API and the DDIApps BFF both build rank image URLs with the same shape.

### 2.1 URL template

```
{ASSET_BASE_URL}/data/roster_ranks/{group}/{rank_id}.jpg?{rank_image}
```

* `ASSET_BASE_URL` — configured via the `ASSET_BASE_URL` environment variable.
* `group` — `floor(rank_id / 1000)`. For the IDs above, this is always `0`.
* `rank_id` — the numeric `rank_id` from `xf_nf_rosters_rank`.
* `rank_image` — the value of `rank.RankImage`, used as a cache buster.

### 2.2 Examples

```text
https://assets.ddi.example/data/roster_ranks/0/1.jpg?1741364618
https://assets.ddi.example/data/roster_ranks/0/22.jpg?24
https://assets.ddi.example/data/roster_ranks/0/31.jpg?31
```

### 2.3 Source code that emits these URLs

* `DDI-API/milpacs/ranks.go` `ImageURL()`
* `CavApps/server/db/milpacDb.js` `rankImageUrl()`

<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/ranks.go" />
<ref_file file="/home/bobbyswagg/research/CavApps/server/db/milpacDb.js" />

## 3. Required Frontend Filenames

`client/app/uniformbuilder/modules/GetUserInfo.jsx` converts `rankId` to `rankGrade`, then `client/app/uniformbuilder/modules/canvas.jsx` loads `client/public/skunkworks/uniformEpaulettes/{rankGrade}.png`.

| rankGrade | filename | source rank_id |
| --- | --- | --- |
| O11 | `O11.png` | 1 |
| O10 | `O10.png` | 2 |
| O9 | `O9.png` | 3 |
| O8 | `O8.png` | 4 |
| O7 | `O7.png` | 5 |
| O6 | `O6.png` | 6 |
| O5 | `O5.png` | 7 |
| O4 | `O4.png` | 8 |
| O3 | `O3.png` | 9 |
| O2 | `O2.png` | 10 |
| O1 | `O1.png` | 11 |
| E11 | `E11.png` | 12 |
| E10 | `E10.png` | 13 |
| E9 | `E9.png` | 14 |
| E8 | `E8.png` | 15 |
| E7 | `E7.png` | 16 |
| E6 | `E6.png` | 17 |
| E5 | `E5.png` | 18 |
| E4 | `E4.png` | 19 |
| E3 | `E3.png` | 20 |
| E2 | `E2.png` | 21 |
| E1 | `E1.png` | 22 |
| E0 | `E0.png` | 23, 31, 32, and any unknown rank_id |
| W5 | `W5.png` | 26 |
| W4 | `W4.png` | 27 |
| W3 | `W3.png` | 28 |
| W2 | `W2.png` | 29 |
| W1 | `W1.png` | 30 |

<ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/canvas.jsx" />

## 4. What Must Not Change

To preserve existing behavior, the following are locked for Phase 5.5A:

* `rank_id` values in the database.
* The `/data/roster_ranks/{group}/{rank_id}.jpg` URL path.
* The `rankImage` cache-buster query parameter.
* The `rankGrade` → `{rankGrade}.png` mapping.
* The existing `client/public/skunkworks/uniformEpaulettes/` filenames.

## 5. DDI Asset Directory Structure (Prepared)

A placeholder directory for future DDI rank artwork has been created. It is not yet used by the application.

```
CavApps/client/public/ddi/rankImages/
  README.md
  .gitkeep
```

When the actual DDI rank art is ready, it can be copied here and `canvas.jsx` can be updated in a later phase. Until then, the uniform builder continues to use `client/public/skunkworks/uniformEpaulettes/`.

## 6. Validation

A non-runtime helper `CavApps/scripts/validate-rank-assets.sh` checks:

* Every `rankGrade` from `GetUserInfo.jsx` has a matching `uniformEpaulettes/*.png` file.
* The DDI placeholder directory exists.

Run with:

```bash
bash scripts/validate-rank-assets.sh
```

The script only prints status and exits with an error code if files are missing. It does not run as part of the build, the server, or any test.