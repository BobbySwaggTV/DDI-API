# DDI Asset Migration Plan

> Phase 5.3 planning: decide how DDI will own the images and sprites that the DDI API and DDIApps currently serve from 7CAV infrastructure.  
> No files were moved, no URLs were changed, and no code was modified.

## 1. Current Asset Sources

### 1.1 DDI API-served assets

The DDI API builds URLs at runtime from `ASSET_BASE_URL` (default `https://7cav.us`) and the path templates below. The files themselves live on the legacy 7CAV asset host unless `ASSET_BASE_URL` is configured.

| Asset class | Go source | Current URL template | Notes |
| --- | --- | --- | --- |
| Rank images | `milpacs/ranks.go` `ImageURL()` | `{ASSET_BASE_URL}/data/roster_ranks/{group}/{rankId}.jpg?{rankImage}` | `group = floor(rankId / 1000)` |
| Uniform images | `milpacs/profiles.go` `UniformUrl()` | `{ASSET_BASE_URL}/data/roster_uniforms/{group}/{relationId}.jpg` | `group = floor(relationId / 1000)` |
| Award images | `milpacs/awards.go` `ImageURL()` | `{ASSET_BASE_URL}/data/roster_awards/{group}/{awardId}.jpg?{awardImage}` | `group = floor(awardId / 1000)` |

These are consumed by:

* `types.Profile.Rank.RankImageUrl`
* `types.Profile.UniformUrl`
* `types.Profile.Awards[].AwardImageUrl`
* `types.RankExpanded.RankImageUrl`

<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/ranks.go" />
<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/profiles.go" />
<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/awards.go" />
<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/asset_url.go" />

### 1.2 DDIApps-bundled assets

The DDIApps client keeps PNG sprite sheets and loose images in `client/public/skunkworks/` and loads them with static paths in `client/app/uniformbuilder/modules/canvas.jsx`.

| Asset bucket | Local path | Usage | Count |
| --- | --- | --- | --- |
| Uniform base | `client/public/skunkworks/uniformBase/` | base uniform, right lapel | 2+ PNGs |
| Rank epaulettes | `client/public/skunkworks/uniformEpaulettes/` | `{rankGrade}.png` (O6, E5, etc.) | ~30 PNGs |
| Ribbons | `client/public/skunkworks/uniformRibbons/ribbons/ribbonSpriteSheet.png` | sprite sheet, 43×14 tiles | 1 sheet |
| Unit citation ribbons | `client/public/skunkworks/uniformRibbons/ribbons/unitCitationSprite.png` | unit citation sprite | 1 sheet |
| Ribbon attachments | `client/public/skunkworks/uniformRibbons/attachments/{type}/{count}.png` | oak leaf, star, etc. | many PNGs |
| Medals | `client/public/skunkworks/uniformMedals/medalSpriteSheet.png` | sprite sheet, 70×120 tiles | 1 sheet |
| Special medals | `client/public/skunkworks/uniformSpecialMedals/{awardPriority}.png` | special single medals | many PNGs |
| Combat badges | `client/public/skunkworks/uniformBadges/combatBadges/{n}.png` | 1–12 CIB/EIB/aviator/etc. | 12 PNGs |
| Tabs | `client/public/skunkworks/uniformTabs/tabSpriteSheet.png` | tab sprite | 1 sheet |
| Weapon qualification plates | `client/public/skunkworks/uniformWeaponQuals/plates/{i}.png` | qualification plates | many PNGs |
| Weapon qualification roots | `client/public/skunkworks/uniformWeaponQuals/root/{i}.png` | base weapon qual images | many PNGs |
| Service stripes | `client/public/skunkworks/uniformService/{enlisted\|officer}/serviceStripe.png` | years-of-service stripes | 2 PNGs |
| Name tags | `client/public/skunkworks/uniformNameTag/{selector}.png` | name-tag textures | many PNGs |
| Shoulder cords | `client/public/skunkworks/uniformCords/{shoulderCord}.png` | MOS-derived cords | many PNGs |
| Lapel pins | `client/public/skunkworks/uniformLapelPins/{neckPins}.png` | MOS-derived pins | many PNGs |
| App logo | `client/public/CavAppsLogo.svg` | homepage logo (unchanged) | 1 SVG |

<ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/canvas.jsx" />

### 1.3 Build-time award assets

`generateAwardSprites.js` reads `uniform-uploads/manifest.json` and `client/app/uniformbuilder/modules/constants/awardCatalog.js`, then slices the uploaded award PNGs into the sprite sheets above.

| Input | Source | Geometry |
| --- | --- | --- |
| Ribbon tiles | `uniform-uploads/` → `client/public/skunkworks/uniformRibbons/ribbons/ribbonSpriteSheet.png` | 43 × 14 px |
| Medal tiles | `uniform-uploads/` → `client/public/skunkworks/uniformMedals/medalSpriteSheet.png` | 70 × 120 px |

The catalog row (`awardPriority` / `medalPriority`) determines the sprite row, so any DDI replacement must preserve the same priority numbering or update the catalog and the coordinate math together.

## 2. Assets Requiring Migration

### 2.1 Must move to DDI ownership

| Asset class | Why | Current owner |
| --- | --- | --- |
| Rank `.jpg` files | API `rankImageUrl` depends on them | 7CAV `https://7cav.us/data/roster_ranks/` |
| Uniform `.jpg` files | API `uniformUrl` depends on them | 7CAV `https://7cav.us/data/roster_uniforms/` |
| Award `.jpg` files | API `awardImageUrl` depends on them; also used in `rosterhistory` timeline | 7CAV `https://7cav.us/data/roster_awards/` |
| DDIApps `skunkworks/` PNGs/SVGs | Client uniform builder literally draws from these files | 7CAV art in `client/public/skunkworks/` |
| App logo SVG | Homepage logo is 7CAV-branded | `client/public/CavAppsLogo.svg` |

### 2.2 Can remain inherited temporarily

| Asset class | Why it can wait | Caveat |
| --- | --- | --- |
| Path templates `/data/roster_ranks/...` | DDI can keep the same URL shape and only change `ASSET_BASE_URL` | The paths themselves are not 7CAV-specific. |
| Sprite-sheet tile geometry (43×14, 70×120) | Generic layout constants | DDI art must either match or `canvas.jsx` / catalog math must change. |
| Rank-grade → filename mapping (`O6.png`, `E5.png`) | Generic US pay grades | If DDI uses different rank IDs, the mapping must change. |

## 3. Recommended DDI Asset Structure

### 3.1 Hosting domain and environment

| Environment | Suggested `ASSET_BASE_URL` | Notes |
| --- | --- | --- |
| Production | `https://assets.ddi.example` | Dedicated DDI asset CDN or bucket. |
| Staging | `https://assets-staging.ddi.example` | Keep separate from prod. |
| Local dev | `http://localhost:port` or the DDI forum host | Optional; the builder can continue to use 7CAV assets locally until DDI art is ready. |

### 3.2 DDI API asset folders

Keep the same path templates so the Go code does not change. Only the `ASSET_BASE_URL` env var changes.

```
{ASSET_BASE_URL}/data/roster_ranks/{group}/{rankId}.jpg?{rankImage}
{ASSET_BASE_URL}/data/roster_uniforms/{group}/{relationId}.jpg
{ASSET_BASE_URL}/data/roster_awards/{group}/{awardId}.jpg?{awardImage}
```

Migration is therefore a bulk copy from the 7CAV host to the new DDI host plus a `find`/`sed` to remove any 7CAV watermarks or filenames if needed.

### 3.3 DDIApps public asset folders

`client/public/skunkworks/` should be mirrored into a `client/public/ddi/` tree (or the filenames can be retained under `skunkworks/` and the art simply replaced). A clean rename is recommended to make the DDI ownership explicit:

```
client/public/
  ddi/
    rankImages/          (epaulettes)
    ribbons/             (sheets + attachments)
    medals/              (sheets + special)
    badges/              (combat badges)
    tabs/                (tab sheet)
    weaponQuals/         (plates + roots)
    serviceStripes/
    nameTags/
    cords/
    lapelPins/
    base/                (uniform base, lapel)
    logo.svg             (DDI logo)
```

If the rename is done, `canvas.jsx` and `AwardClasses.jsx` will need updates in a later phase to point at the new paths. If files are simply replaced in place, no code changes are required for the uniform builder beyond `awardCatalog.js`.

### 3.4 Naming conventions

| Asset type | Convention | Example |
| --- | --- | --- |
| Rank images | `{group}/{rankId}.jpg` with cache-buster query | `data/roster_ranks/0/5.jpg?12345` |
| Uniform images | `{group}/{relationId}.jpg` | `data/roster_uniforms/0/105.jpg` |
| Award images | `{group}/{awardId}.jpg` with cache-buster query | `data/roster_awards/0/1.jpg?12345` |
| Epaulettes | `{rankGrade}.png` | `O6.png` |
| Ribbons | `ribbonSpriteSheet.png` | sprite sheet |
| Medals | `medalSpriteSheet.png` | sprite sheet |
| Special medals | `{awardPriority}.png` | `5.png` |
| Combat badges | `{n}.png` | `7.png` |
| Weapon qual plates | `{qualificationKey}.png` | `expert.png` |
| Cords / pins | `{mosGroupKey}.png` | `11B.png` |

The new convention should keep the numeric keys used by `awardCatalog.js` (`awardPriority`, `medalPriority`) and `GetUserInfo.jsx` (`rankGrade`, `shoulderCord`, `neckPins`) stable to avoid code churn.

## 4. Migration Order

1. **Low risk — set `ASSET_BASE_URL` to DDI host (Phase 5.3.1)**
   - Copy the existing 7CAV rank, uniform, and award `.jpg` files to `assets.ddi.example`.
   - Update the DDI API `ASSET_BASE_URL` and the DDIApps server `ASSET_BASE_URL`.
   - The API URLs become DDI-owned immediately without code changes.

2. **Medium risk — replace rank epaulettes and base uniform art (Phase 5.3.2)**
   - Replace `client/public/skunkworks/uniformEpaulettes/` and `uniformBase/` PNGs with DDI art.
   - Keep filenames identical (e.g. `O6.png`) and the same image dimensions.
   - If DDI rank grades differ, update `GetUserInfo.jsx` `getRankGrade()`.

3. **Medium risk — replace ribbon/medal sheets (Phase 5.3.3)**
   - Regenerate `ribbonSpriteSheet.png` and `medalSpriteSheet.png` from `uniform-uploads/`.
   - Keep the 43×14 and 70×120 tile sizes.
   - Keep `awardCatalog.js` row priorities stable or migrate them with the sprites.

4. **High risk — rebuild the DDI award catalog and sprites (Phase 5.3.4)**
   - Create DDI `awardCatalog.js` entries with new award names, priorities, types, and attachment types.
   - Add the matching award PNGs to `uniform-uploads/`.
   - Run `npm run sprites:generate`.
   - Update the DDI API database `xf_nf_rosters_award.title` so `AwardName` matches the catalog keys exactly.

5. **High risk — replace remaining builder assets and logos (Phase 5.3.5)**
   - Combat badges, tabs, weapon quals, service stripes, name tags, cords, lapel pins, and the app logo.
   - Update `client/app/theme/appsLogo.js` or replace `client/public/CavAppsLogo.svg`.

6. **Cleanup — remove 7CAV default and test fixtures (Phase 5.3.6)**
   - Change `milpacs/asset_url.go` `defaultAssetBaseURL` to a DDI placeholder or make it error if unset.
   - Update `generateAwardSprites.js` if the output paths move from `skunkworks/` to `ddi/`.

## 5. Risks and Mitigations

| Risk | When it happens | Mitigation |
| --- | --- | --- |
| **Broken rank/uniform/award URLs** | `ASSET_BASE_URL` is set before the DDI CDN has the files, or the files are not at the expected path. | Stage assets first; validate with a dry-run of the BFF cache. Keep `ASSET_BASE_URL` unset during the copy, then switch over atomically. |
| **Award matching failure** | DDI award titles in `xf_nf_rosters_award.title` do not match `awardCatalog.js` keys. | Make `awardCatalog.js` the source of truth and use it to seed award titles. If the DB is upstream, add a mapping file for old→new award names. |
| **Sprite coordinate drift** | New DDI ribbon/medal art uses different dimensions or the catalog priorities are reordered. | Freeze the 43×14 and 70×120 tile sizes. If priorities change, regenerate the sheets and verify with `npm run test:sprites` if available. |
| **Cache behavior** | Browser/CDN caches old 7CAV images after the URL switch. | Add cache-buster query parameters to the API URLs (already supported by `?rankImage`, `?awardImage`). Use `next/image` headers or versioned filenames for sprite sheets. |
| **Sprite-sheet generation failure** | `generateAwardSprites.js` fails because a new DDI award does not fit the catalog. | Add awards one at a time, run the generator, and keep the `uniform-uploads/manifest.json` in version control. |
| **Rank ID renumbering** | DDI renumbers `rank_id` values in the database. | Keep the old `rank_id` values until `GetUserInfo.jsx` `getRankGrade()` and the epaulette PNG names are remapped. |
| **MOS-driven cords/pins** | DDI MOS strings differ from `11B`, `68W`, etc. | Map DDI MOS codes to the existing cord/pin PNG names or keep the same codes. |

## 6. Summary

* The DDI API only needs `ASSET_BASE_URL` to change; the URL templates can remain.
* The DDIApps client needs the actual PNG/SVG art in `client/public/skunkworks/` (or a new `ddi/` folder) replaced.
* The award catalog and sprite sheets are the highest-risk migration because the uniform builder silently skips awards it does not recognize.
* The safest order is: set the DDI asset host → replace rank/base art → regenerate ribbon/medal sheets → rebuild the award catalog → replace remaining badges/tabs/logo → clean up legacy defaults.