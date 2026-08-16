# DDI Award Asset Migration Plan

> Phase 5.7B: plan how to move the award sprite sheets, uploaded PNG sources, and catalog metadata to DDI ownership.  
> No sprite sheets were rebuilt, no awards were renamed, no catalog entries were modified, and no API contracts were changed.

## 1. Current Award Asset System

### 1.1 Sprite sheets

| Sheet | File | Tile size | Row origin | Catalog field |
| --- | --- | --- | --- | --- |
| Ribbons | `client/public/skunkworks/uniformRibbons/ribbons/ribbonSpriteSheet.png` | 43 × 14 px | `y = awardPriority * 14` | `awardPriority` |
| Medals | `client/public/skunkworks/uniformMedals/medalSpriteSheet.png` | 70 × 120 px | `y = (medalPriority - 2) * 120` | `medalPriority` |

Both sheets are opaque PNGs with the tiles packed vertically. The uniform builder samples a row based on the `awardPriority` or `medalPriority` from the catalog.

### 1.2 Source upload workflow

```
uniform-uploads/manifest.json
  ↓
generateAwardSprites.js
  ↓
client/public/skunkworks/uniformRibbons/ribbons/ribbonSpriteSheet.png
client/public/skunkworks/uniformMedals/medalSpriteSheet.png
```

`generateAwardSprites.js`:

1. Reads `uniform-uploads/manifest.json`.
2. Validates each entry against `awardCatalog.js`.
3. Normalizes each source PNG to the correct tile size.
4. Slices it into the row determined by `awardPriority` or `medalPriority`.
5. Writes the updated sprite sheets.

<ref_file file="/home/bobbyswagg/research/CavApps/generateAwardSprites.js" />

### 1.3 Canvas coordinate references

`client/app/uniformbuilder/modules/canvas.jsx` references the sheets with relative paths:

```javascript
"skunkworks/uniformRibbons/ribbons/ribbonSpriteSheet.png"
"skunkworks/uniformMedals/medalSpriteSheet.png"
```

The `getCanvasObject.jsx` (or similar helper) computes the row from the catalog values. No coordinate math should change during an asset migration.

### 1.4 Attachment devices

Attachment PNGs live outside the main sheets:

```
client/public/skunkworks/uniformRibbons/attachments/
  oakClusters/
  oakClustersService/
  oakClustersValor/
  stars/
  starsDonation/
  silverStars/
  ncoNums/
  gcNotches/
```

Each `awardAttachmentType` in the catalog selects a folder and a `MAX_AWARD_COUNT` cap.

<ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/constants/awardAttachmentTypes.js" />

### 1.5 API-served award images

The DDI API emits `AwardImageUrl` from:

```go
fmt.Sprintf("%s/data/roster_awards/%d/%d.jpg?%d",
  assetBaseURL(), imageGroup, award.AwardId, award.AwardImage)
```

These are portrait/award images shown on the roster, not the sprite sheets used by the uniform builder.

<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/awards.go" />

## 2. Current Catalog Dependencies

### 2.1 `awardPriority`

* Determines the ribbon row.
* Must be an integer `>= 0`.
* Insertion order in `awardCatalog.js` is the current visual order.

### 2.2 `medalPriority`

* Determines the medal row.
* Used only for `Medal`, `MedalWithValor`, and `MedalTiered` types.
* The first two rows (`0`, `1`) are for Lifetime medals, which live on the ribbon sheet only, not the medal sheet.

### 2.3 `awardAttachmentType`

* Optional.
* One of: `OAK_CLUSTERS`, `OAK_CLUSTERS_SERVICE`, `OAK_CLUSTERS_VALOR`, `SILVER_STARS`, `STARS`, `STARS_DONATION`, `GC_NOTCHES`, `NCO_NUMS`.
* Drives which attachment folder is loaded and how many can be rendered.

### 2.4 `awardType`

* One of: `Medal`, `MedalWithValor`, `MedalTiered`, `Ribbon`, `RibbonDonationLogic`, `UnitCitation`, `BadgeCombat`, `WeaponQual`, `Tab`.
* Determines which sheet (if any) the award uses and which class (`Award`, `MedalWithValor`, `MedalTiered`, `Ribbon`, `BadgeCombat`, etc.) renders it.

## 3. Recommended DDI Migration Strategy

### 3.1 Phase 1: DDI host for API award images (low risk)

1. Copy existing 7CAV `data/roster_awards/{group}/{awardId}.jpg` files to the DDI asset host.
2. Set `ASSET_BASE_URL` to the DDI host in both the DDI API and DDIApps BFF.
3. The roster award portraits now come from DDI infrastructure.
4. The uniform builder is not touched.

### 3.2 Phase 2: preserve official award names (zero risk)

* Do not rename any `name` in `awardCatalog.js`.
* Use `client/app/uniformbuilder/modules/constants/awardCompatibility.mjs` (Phase 5.7A) to map any new DDI display names to the existing catalog keys.
* Official award identities are unchanged.

### 3.3 Phase 3: replace DDI-specific award art in place (medium risk)

1. Create DDI artwork for the 7CAV-specific awards.
2. Place the new PNGs in `uniform-uploads/` with the same manifest row and `awardId` mapping.
3. Run `npm run sprites:generate` to rebuild only the affected rows.
4. Keep the same `awardPriority` / `medalPriority` values, so `canvas.jsx` coordinates do not change.

### 3.4 Phase 4: add new DDI awards (high risk)

1. Append new entries at the end of `awardCatalog.js` with the next available `awardPriority` and `medalPriority` values.
2. Do **not** reuse or reshuffle existing priorities.
3. Add new PNGs to `uniform-uploads/` and run `npm run sprites:generate`.
4. Update `awardCompatibility.mjs` aliases if the API uses different DDI names.

### 3.5 Phase 5: DDI-specific attachment artwork (medium risk)

1. Replace PNGs in `uniformRibbons/attachments/{device}/`.
2. Keep the same filenames (`1.png`, `2.png`, etc.).
3. Do not change `MAX_AWARD_COUNT` unless DDI changes the display rule.

## 4. Sprite Rebuild Process

### 4.1 Before rebuilding

1. Back up the existing sheets and `uniform-uploads/`.
2. Verify the manifest matches the catalog.
3. Ensure the new source PNGs are at least as large as the tile size (43×14 or 70×120). The generator normalizes them.

### 4.2 Rebuild command

```bash
npm run sprites:generate
```

### 4.3 After rebuilding

1. Run `scripts/validate-awards.mjs` (official names and aliases still resolve).
2. Visually inspect the new `ribbonSpriteSheet.png` and `medalSpriteSheet.png`.
3. Render a sample uniform for each `awardType` and `awardAttachmentType`.

## 5. Rollback Strategy

1. **API images** — revert `ASSET_BASE_URL` to the previous host (e.g. `https://7cav.us`).
2. **Sprite sheets** — restore `ribbonSpriteSheet.png` and `medalSpriteSheet.png` from Git or backup.
3. **Catalog** — revert `awardCatalog.js` to the previous commit.
4. **Aliases** — revert `awardCompatibility.mjs` or remove the new aliases.

Because `awardCatalog.js` `name` values are not renamed, the builder can always fall back to the previous sheet as long as the priorities match.

## 6. Testing Requirements

| Test | When |
| --- | --- |
| `npm run sprites:generate` completes without validation errors | Before every sprite change |
| `scripts/validate-awards.mjs` passes | After alias or catalog changes |
| `go test ./milpacs` passes | After API asset URL changes |
| `npm --prefix server test` passes | After BFF changes |
| Render a sample uniform for at least one of each `awardType` | After sprite rebuild |
| Render a sample with an `OAK_CLUSTERS`, `STARS`, and `NCO_NUMS` attachment | After attachment art changes |

## 7. Handling New Awards

1. **Assign new priorities** at the end of the current ranges.
2. **Append to `awardCatalog.js`**, keeping the original entries untouched.
3. **Place source PNGs** in `uniform-uploads/` and add them to `manifest.json`.
4. **Run the generator**.
5. **Update `awardCompatibility.mjs`** if the API title differs from the catalog `name`.
6. **Test** that `AwardRegistry.getAwardDetails(newName)` returns a valid object and `canvas.jsx` draws the correct row.

## 8. Summary

* The safest migration is to first move the API-served award `.jpg` files to a DDI host, then replace the sprite artwork in place, then add new DDI awards at the end of the catalog.
* Never reuse existing `awardPriority` or `medalPriority` values for a different award.
* Never rename official award catalog `name` values; use `awardCompatibility.mjs` for display-name mapping.
* Keep tile geometry (43×14, 70×120) and `MAX_AWARD_COUNT` stable unless a deliberate, tested UI change is required.