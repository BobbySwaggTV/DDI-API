# DDI Award Asset Production Specification

> Phase 5.7E: production handoff for DDI award graphics.  
> No awards were renamed, no catalog priorities were changed, no API contracts were modified, and no sprites were rebuilt.

This document tells artists and developers exactly how to produce new award assets for the DDIApps uniform builder so they can be dropped into the existing sprite-sheet pipeline without breaking the current rendering logic.

## 1. Ribbon Asset Requirements

### 1.1 Dimensions

* **Tile size:** 43 × 14 px
* **Source asset:** at least 43 × 14 px (the generator will normalize oversized assets)
* **Aspect ratio:** 43:14
* **Mode:** RGBA PNG with transparency

### 1.2 Transparency

* Ribbons are full-width rectangular bands.
* The background must be transparent so they can be layered on the uniform template.
* Do not use a solid color fill outside the ribbon design.

### 1.3 Naming convention

Place the source PNG in `uniform-uploads/` and add it to `uniform-uploads/manifest.json`.

```json
{
  "name": "DDI Service Ribbon",
  "awardPriority": 105,
  "medalPriority": null,
  "replace": false
}
```

* `name` — must match the `awardCatalog.js` `name` exactly (or be mapped through `awardCompatibility.mjs`).
* `awardPriority` — the catalog row where the ribbon tile belongs.
* `medalPriority` — `null` if the award has no medal tile.
* `replace` — `true` only when overwriting existing art in place.

### 1.4 Sprite placement

* The generator inserts the tile at row `awardPriority * 14`.
* Keep `awardPriority` an integer `>= 0`.
* New DDI ribbons must use the next free `awardPriority` at the end of the sheet.
* Do not reuse an existing `awardPriority` for a different award.

## 2. Medal Asset Requirements

### 2.1 Dimensions

* **Tile size:** 70 × 120 px
* **Source asset:** at least 70 × 120 px
* **Aspect ratio:** 7:12
* **Mode:** RGBA PNG with transparency

### 2.2 Transparency

* Medals hang from a suspension ribbon at the top.
* The top of the tile should include the suspension ribbon and/or ring.
* The background must be transparent.

### 2.3 Sprite placement

* Medals sit on a separate sprite sheet from ribbons.
* The first two rows (`medalPriority` 0 and 1) are reserved for Lifetime medals that appear only on the ribbon sheet.
* All other medals start at `medalPriority = 2`.
* Row `y` is calculated as `(medalPriority - 2) * 120`.
* New DDI medals must use the next free `medalPriority >= 2`.

### 2.4 Manifest example

```json
{
  "name": "DDI Lifetime Dedication Medal",
  "awardPriority": 0,
  "medalPriority": 2,
  "replace": false
}
```

If the medal has no ribbon tile, set `awardPriority` to `null`.

## 3. Badge / Tab Requirements

### 3.1 Combat badges

* Rendered from `client/public/skunkworks/uniformBadges/combatBadges/{n}.png`
* `n` is computed at runtime from `awardPriority` via `AwardClasses.jsx` `getImageNum()`.
* File dimensions and format are not rigid (the canvas draws the PNG at a fixed coordinate), but the artwork should fit the standard badge dimensions used by the current 12 images.
* For DDI-specific badges, place new PNGs in `client/public/skunkworks/uniformBadges/combatBadges/` and update `client/app/uniformbuilder/modules/constants/badgeImages.js` with the new image-number mapping.

### 3.2 Tabs

* Rendered from `client/public/skunkworks/uniformTabs/tabSpriteSheet.png`
* Tabs are not produced through `generateAwardSprites.js`.
* New DDI tabs should be placed in a separate `client/public/ddi/uniformTabs/` directory or added to the existing sprite sheet.
* If a new tab type is introduced, append it to `awardCatalog.js` with `awardType: AwardType.Tab` and add a new coordinate entry in `canvas.jsx` or the tab helper.

## 4. Attachment / Device Requirements

Attachment devices are layered on top of ribbon/medal tiles. They are stored as small PNGs in `client/public/skunkworks/uniformRibbons/attachments/`.

| Attachment type | Folder | Device dimensions (typical) | Max count |
| --- | --- | --- | --- |
| Oak leaf clusters | `oakClusters/` | small square, fits on ribbon | 19 |
| Service oak leaf clusters | `oakClustersService/` | small square | 6 |
| Valor oak leaf clusters | `oakClustersValor/` | small square | 14 |
| Campaign stars | `stars/` | small star | 10 |
| Donation stars | `starsDonation/` | small star | 12 |
| Silver stars | `silverStars/` | small star | 5 |
| NCO numerals | `ncoNums/` | small numeral | 6 |
| Good conduct notches | `gcNotches/` | small bar | 9 |

### 4.1 File naming

Each folder contains sequential files: `1.png`, `2.png`, `3.png`, etc. The number corresponds to the number of devices displayed.

### 4.2 Replacement

For a DDI rebrand, replace the existing `1.png`, `2.png`, etc. with DDI-styled devices. Keep the same dimensions so the canvas coordinate math does not change.

## 5. File Organization for DDI-Owned Award Assets

### 5.1 Recommended source tree

```
CavApps/
  uniform-uploads/
    ribbons/           # 43x14 source PNGs
    medals/            # 70x120 source PNGs
    badges/            # combat badge source PNGs
    attachments/       # device source PNGs
    manifest.json
  client/public/
    skunkworks/        # legacy 7CAV assets (kept as fallback)
    ddi/               # new DDI assets (future ownership)
      uniformRibbons/ribbons/ribbonSpriteSheet.png
      uniformMedals/medalSpriteSheet.png
      uniformBadges/combatBadges/
      uniformRibbons/attachments/
```

### 5.2 Migration in place (recommended)

During the initial migration, DDI can replace files inside `client/public/skunkworks/` without changing the paths in `canvas.jsx`. This is the lowest-risk path.

### 5.3 Long-term ownership

After the initial migration, DDI should move the assets to `client/public/ddi/` and update the builder paths. This can be done incrementally per award category.

## 6. Sprite Generation Workflow

### 6.1 Source assets

1. Create or acquire a PNG for each award.
2. Place it under `uniform-uploads/` in the correct `ribbons/`, `medals/`, `badges/`, or `attachments/` folder.
3. Add an entry to `uniform-uploads/manifest.json`.

### 6.2 Generation script

Run the generator to splice the source PNGs into the sprite sheets:

```bash
npm run sprites:generate
```

<ref_file file="/home/bobbyswagg/research/CavApps/generateAwardSprites.js" />

### 6.3 Validation steps

The generator validates:

* Every manifest entry has a matching `awardCatalog.js` `name`.
* `awardPriority` / `medalPriority` are valid row indices.
* Source PNGs match the expected ribbon or medal dimensions.
* The catalog and sheet row counts are consistent.

## 7. Testing Checklist

### 7.1 Catalog lookup

- [ ] Run `node scripts/validate-awards.mjs` to confirm official names and aliases resolve.
- [ ] For any new DDI award, add its API-facing name to `awardCompatibility.mjs` and re-run the script.
- [ ] Verify `AwardRegistry.getAwardDetails(newAwardName)` returns a non-zero object.

### 7.2 Sprite generation

- [ ] Run `npm run sprites:generate` and ensure it completes with no manifest errors.
- [ ] Open the generated `ribbonSpriteSheet.png` and confirm the new tile is at the correct row.
- [ ] Open the generated `medalSpriteSheet.png` and confirm the new medal is at the correct row.
- [ ] Restore `uniform-uploads/` and any replaced PNGs from the manifest after verification.

### 7.3 Priority placement

- [ ] Confirm the new `awardPriority` is at the end of the current ribbon rows.
- [ ] Confirm the new `medalPriority` is at the end of the current medal rows and `>= 2`.
- [ ] Confirm no existing `awardPriority` / `medalPriority` was reshuffled.

### 7.4 Uniform output validation

- [ ] Render a sample uniform for a member with the new award.
- [ ] Verify the ribbon appears at the expected position on the uniform.
- [ ] If the award has a medal, verify the medal is rendered correctly.
- [ ] If the award has an attachment device, verify the correct device PNG is layered on top.
- [ ] Run `npm --prefix server test` and `node server/controllers/diffEngine.test.js` to ensure BFF compatibility.

## 8. Rules Summary

| Rule | Why |
| --- | --- |
| Do not rename official awards | `awardCatalog.js` is the authority; `AwardRegistry` looks up by exact name. |
| Do not change catalog priorities | `canvas.jsx` row math depends on `awardPriority` and `medalPriority`. |
| Do not modify API contracts | The DDI API `Award` JSON shape must remain stable. |
| Do not rebuild sprites until assets are ready | Rebuilding before DDI art exists would corrupt the sprite sheets. |

## 9. Handoff Summary

1. Artists produce 43×14 ribbon PNGs, 70×120 medal PNGs, and optional badge/tab PNGs with transparent backgrounds.
2. Developers add the assets to `uniform-uploads/` and the manifest.
3. Developers run `npm run sprites:generate`.
4. QA runs `scripts/validate-awards.mjs`, `npm --prefix server test`, and a sample uniform render.
5. If everything passes, the new DDI assets are committed; the catalog names and priorities remain unchanged.