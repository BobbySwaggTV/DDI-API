# DDI Award Migration Audit

> Phase 5.7: identify 7CAV-specific award content and its impact on the uniform builder.  
> No code, award names, catalog entries, sprite sheets, or API contracts were changed.

## 1. Current Award Data Flow

### 1.1 DDI API → BFF → Frontend

```
xf_nf_rosters_award.title
  ↓
DDI API /milpacs/profiles.go → types.Award.AwardName
  ↓
DDI API JSON response (AwardName, AwardDetails, AwardDate, AwardImageUrl, AwardUid)
  ↓
CavApps BFF passes the array through to the uniform builder
  ↓
AwardRegistry.getAwardDetails(awardData.awardName)
  ↓
canvas.jsx renders the ribbon/medal/tab/badge
```

The DDI API builds `AwardImageUrl` with:

```go
fmt.Sprintf("%s/data/roster_awards/%d/%d.jpg?%d",
  assetBaseURL(), imageGroup, award.AwardId, award.AwardImage)
```

<ref_file file="/home/bobbyswagg/research/DDI-API/milpacs/awards.go" />
<ref_file file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" />
<ref_file file="/home/bobbyswagg/research/DDI-API/types/profiles.go" />

### 1.2 DDIApps award registry

`AwardRegistry.jsx` loads `AWARD_CATALOG` into a `Map` keyed by award name. At render time the builder asks:

1. `isInRegistry(awardName)` — does the catalog know this award?
2. `getAwardDetails(awardName)` — what are the `awardPriority`, `medalPriority`, `awardType`, `awardAttachmentType`?
3. `getMaxAwardCount(awardName)` — how many attachment devices can be displayed?

If an API award name is not in the catalog, the builder silently skips it.

<ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/AwardRegistry.jsx" />

## 2. Current Award Catalog

`client/app/uniformbuilder/modules/constants/awardCatalog.js` contains ~75 award entries. The catalog is ordered by `awardPriority`. Each entry defines:

| Field | Meaning |
| --- | --- |
| `name` | Exact string matched against `AwardName` from the API. |
| `awardPriority` | Vertical row in `ribbonSpriteSheet.png` (43×14 px tiles). |
| `medalPriority` | Vertical row in `medalSpriteSheet.png` (70×120 px tiles). |
| `awardType` | One of `Medal`, `MedalWithValor`, `MedalTiered`, `Ribbon`, `RibbonDonationLogic`, `UnitCitation`, `BadgeCombat`, `WeaponQual`, `Tab`. |
| `awardAttachmentType` | How oak leaves/stars/etc. are drawn (optional). |

### 2.1 Generic awards

| Name | Type | Notes |
| --- | --- | --- |
| `Army Distinguished Service Cross` | Medal | U.S. Army decoration |
| `Defense Distinguished Service Medal` | Medal | U.S. DoD decoration |
| `Silver Star` | Medal | U.S. decoration |
| `Bronze Star` | MedalWithValor | U.S. decoration |
| `Purple Heart` | Medal | U.S. decoration |
| `Army Commendation Medal` | MedalWithValor | U.S. decoration |
| `Army Achievement Medal` | Medal | U.S. decoration |
| `Prisoner of War Medal` | Medal | U.S. decoration |
| `Air Medal` | Medal | U.S. decoration |
| `Global War on Terrorism Expeditionary Medal` | Medal | U.S. campaign medal |
| `National Defense Service Medal` | Medal | U.S. service medal |
| ... | ... | All standard U.S. Army/DoD awards |

### 2.2 7CAV-specific awards

| Name | Type | Notes |
| --- | --- | --- |
| `7th Cavalry Lifetime Dedication Award` | Medal | 7CAV unit-specific |
| `7th Cavalry Server Upgrade Award` | MedalTiered | 7CAV unit-specific |
| `7th Cavalry Black Ops Unit Citation` | UnitCitation | 7CAV unit-specific |
| `Cavalry Centurion Medal` | Medal | 7CAV / cavalry-specific |
| `Ready or Not Service Ribbon` | Medal | 7CAV game server |
| `DCS World Service Ribbon` | Medal | 7CAV game server |
| `Squad Service Ribbon` | Medal | 7CAV game server |
| `Hell Let Loose Service Ribbon` | Medal | 7CAV game server |
| `Hell Let Loose Console Service Ribbon` | Medal | 7CAV game server |
| `Battlefield 6 Service Ribbon` | Medal | 7CAV game server |
| `Foxhole Service Ribbon` | Medal | 7CAV game server |
| `James "Krazee" Foster Lifetime Achievement Medal` | Medal | 7CAV-named |
| `Ronnie "Coldblud" Bussey Lifetime Achievement Medal` | Medal | 7CAV-named |

<ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/constants/awardCatalog.js" />

## 3. Exact-Name Matching

### 3.1 Registry lookup

```javascript
// AwardRegistry.jsx
getAwardDetails(awardName) {
  if (hasValorDevice(awardName)) {
    awardName = stripValorDevice(awardName);
  }
  return this.awards.get(awardName) ?? 0;
}
```

* The only normalization is stripping the `" with Valor Device"` suffix.
* Otherwise the match is case-sensitive and exact.
* A mismatch causes the award to be skipped, which is the highest risk in a DDI migration.

### 3.2 Valor-device handling

`awardNames.js`:

```javascript
export const AwardNameFragment = Object.freeze({
  VALOR_DEVICE: "with Valor Device",
  FLIGHT_MEDIC_BADGE: "Flight Medic Badge",
  AVIATOR: "Aviator",
});
```

If the API title ends with `" with Valor Device"`, the suffix is removed before the catalog lookup so the base medal entry can be used.

<ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/constants/awardNames.js" />

## 4. Sprite Sheet Dependencies

### 4.1 Geometry

| Sheet | File | Tile size | Row calculation |
| --- | --- | --- | --- |
| Ribbons | `client/public/skunkworks/uniformRibbons/ribbons/ribbonSpriteSheet.png` | 43×14 | `y = awardPriority * 14` |
| Medals | `client/public/skunkworks/uniformMedals/medalSpriteSheet.png` | 70×120 | `y = (medalPriority - 2) * 120` |

### 4.2 Generator

`generateAwardSprites.js` at the repo root reads `uniform-uploads/manifest.json` and slices award PNGs into the sheets. It validates the manifest against `awardCatalog.js` because `awardPriority` and `medalPriority` determine the target row.

<ref_file file="/home/bobbyswagg/research/CavApps/generateAwardSprites.js" />

### 4.3 Attachment sprites

Attachment devices are not in the main sprite sheets. They are small PNGs in:

* `client/public/skunkworks/uniformRibbons/attachments/oakClusters/`
* `client/public/skunkworks/uniformRibbons/attachments/stars/`
* `client/public/skunkworks/uniformRibbons/attachments/ncoNums/`
* ...

The attachment type comes from `awardAttachmentType` in the catalog.

## 5. Attachment / Device Dependencies

`awardAttachmentTypes.js` defines the taxonomy and the max number of devices each can display.

| Attachment type | Folder | Max count |
| --- | --- | --- |
| `OAK_CLUSTERS` | `oakClusters/` | 19 |
| `OAK_CLUSTERS_SERVICE` | `oakClusters/` | 6 |
| `OAK_CLUSTERS_VALOR` | `oakClusters/` | 14 |
| `STARS` | `stars/` | 10 |
| `STARS_DONATION` | `stars/` | 12 |
| `SILVER_STARS` | `silverStars/` | 5 |
| `NCO_NUMS` | `ncoNums/` | 6 |
| `GC_NOTCHES` | `gcNotches/` | 9 |

These device types and counts are generic and can remain. The actual PNG artwork can be replaced.

<ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/constants/awardAttachmentTypes.js" />

## 6. DDI Award Migration Strategy

### 6.1 Lowest-risk: API `AwardImageUrl`

Set `ASSET_BASE_URL` to a DDI host and copy the 7CAV award `.jpg` files to:

```
{ASSET_BASE_URL}/data/roster_awards/{group}/{awardId}.jpg?{awardImage}
```

No catalog, sprite sheet, or name changes are required.

### 6.2 Medium-risk: rename awards in the catalog

1. Rename the DDI-specific awards in `awardCatalog.js` (e.g. replace `7th Cavalry` with `DDI`).
2. Update `xf_nf_rosters_award.title` or add a DDI→legacy name mapping in the BFF.
3. If the API titles are updated first, every renamed award must also be renamed in the catalog, or it will not render.

### 6.3 High-risk: replace DDI-specific award art

1. Add new `awardPriority` / `medalPriority` rows to the catalog.
2. Place new PNGs in `uniform-uploads/`.
3. Run `npm run sprites:generate` to rebuild `ribbonSpriteSheet.png` and `medalSpriteSheet.png`.
4. Keep the same tile geometry and priority spacing.

### 6.4 Highest-risk: DDI creates a new award list

If DDI does not use the 7CAV award list at all, the safest approach is:

1. Keep the 7CAV awards in the catalog (for existing profiles).
2. Add new DDI awards at the end of the catalog with new priorities.
3. Map API award names to the new catalog names via a BFF middleware.
4. Rebuild sprite sheets with both sets of tiles.

## 7. Risks

| Risk | Impact | Mitigation |
| --- | --- | --- |
| **Award-name mismatch** | API `AwardName` does not match `awardCatalog.js` `name`. The award is silently skipped. | Keep a name-mapping layer or update catalog and database simultaneously. |
| **Sprite priority collision** | A new DDI award reuses an existing `awardPriority` or `medalPriority`. The wrong tile renders. | Reserve new priorities at the end of the sheet. Use `replace: true` only to overwrite existing art in place. |
| **Missing attachment art** | DDI award uses an attachment type not in `uniformRibbons/attachments/`. | Verify every `awardAttachmentType` has a matching PNG sequence. |
| **Medal/Ribbon row math** | `canvas.jsx` row calculations depend on the 43×14 / 70×120 geometry. | Freeze tile sizes when regenerating sheets. |
| **Valor-device drift** | DDI award names include a different valor suffix. | Add the suffix to `AwardNameFragment` and `hasValorDevice` / `stripValorDevice`. |
| **MAX_AWARD_COUNT drift** | A DDI attachment has a different maximum displayable count. | Update `MAX_AWARD_COUNT` for the relevant `AwardAttachmentType`. |
| **API contract breakage** | Changing `types.Award` JSON shape. | Keep `AwardName`, `AwardDetails`, `AwardDate`, `AwardImageUrl`, `AwardUid`. |

## 8. Rollback Plan

1. **Image URL rollback** — revert `ASSET_BASE_URL` to `https://7cav.us` (or the previous DDI host). The `AwardImageUrl` path templates are the same.
2. **Catalog-name rollback** — revert `awardCatalog.js` to the previous names, or keep a name-mapping file.
3. **Sprite sheet rollback** — restore `ribbonSpriteSheet.png` and `medalSpriteSheet.png` from version control.
4. **BFF name-mapping rollback** — remove any new name-translation middleware so the original names are sent to the builder unchanged.

## 9. Summary

* The DDI API already provides `AwardName` as a free-text title. The only controlled field is `AwardImageUrl`.
* The uniform builder depends on exact match of `AwardName` against `awardCatalog.js`.
* Most U.S. Army awards are generic; 7CAV-specific named awards and game-server ribbons are not.
* Sprite-sheet geometry (43×14 ribbons, 70×120 medals) and attachment counts are the most fragile invariants.
* The safest DDI migration is: set `ASSET_BASE_URL`, then rename awards with a catalog + database lockstep, then regenerate art.