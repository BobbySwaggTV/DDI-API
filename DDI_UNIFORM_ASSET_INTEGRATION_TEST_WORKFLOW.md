# DDI Uniform Asset Integration Test Workflow

> Phase 5.8: safe testing for DDI-owned uniform assets.  
> No existing artwork was replaced, no sprites were rebuilt, no award names were renamed, no catalog priorities were changed, and no API contracts were modified.

This document describes how to load, configure, and validate DDI uniform assets in a local or staging environment without touching the production 7CAV artwork.

## 1. Loading DDI Assets Locally

### 1.1 Rank images

The DDI API and BFF build rank image URLs from `ASSET_BASE_URL`:

```text
{ASSET_BASE_URL}/data/roster_ranks/{group}/{rank_id}.jpg?{rank_image}
```

For local testing, set `ASSET_BASE_URL` to a DDI staging bucket or a `file://` host.

<ref_file file="/home/bobbyswagg/research/DDI-API/DDI_RANK_ASSET_REQUIREMENTS.md" />

### 1.2 Epaulettes (client)

The uniform builder uses `client/public/ddi/rankImages/{rankGrade}.png` first, then falls back to `client/public/skunkworks/uniformEpaulettes/{rankGrade}.png`.

### 1.3 Cords and pins

`mosConfig.mjs` maps MOS codes to cord/pin names. The canvas loads:

* `client/public/skunkworks/uniformCords/{cord}.png`
* `client/public/skunkworks/uniformLapelPins/{pin}.png`

For DDI testing, place DDI cord/pin PNGs in `client/public/ddi/uniformCords/` and `client/public/ddi/uniformLapelPins/`, then update `canvas.jsx` to load the DDI path.

### 1.4 Awards

Awards render from the existing sprite sheets. DDI source PNGs are placed in `uniform-uploads/` and spliced by `npm run sprites:generate`. Keep the original sheets backed up.

<ref_file file="/home/bobbyswagg/research/DDI-API/DDI_AWARD_ASSET_PRODUCTION_SPEC.md" />

## 2. Configuring `ASSET_BASE_URL` for Testing

### 2.1 DDI API

In `docker-compose.yaml`:

```yaml
environment:
  ASSET_BASE_URL: ${ASSET_BASE_URL}
```

For a staging stack:

```bash
ASSET_BASE_URL=https://assets-staging.ddi.example docker compose up
```

For a dev stack, leave it unset; the API falls back to `https://7cav.us`.

### 2.2 DDIApps BFF

In `.env`:

```text
ASSET_BASE_URL=https://assets-staging.ddi.example
```

In `docker-compose.yml`:

```yaml
environment:
  ASSET_BASE_URL: ${ASSET_BASE_URL:-https://7cav.us}
```

For local dev with the legacy assets, leave `.env` empty or use the default.

### 2.3 Static assets

For client-side sprite sheets, the paths are hardcoded relative to `client/public/`. DDI assets can be tested by:

1. Replacing files in `client/public/skunkworks/` (in-place replacement).
2. Or adding `client/public/ddi/` files and updating the import path in `canvas.jsx`.

## 3. Validation

### 3.1 Ranks

Run `scripts/validate-rank-assets.sh`:

```bash
bash scripts/validate-rank-assets.sh
```

It checks:

* Every `rankGrade` has a legacy `uniformEpaulettes/*.png` file.
* The `client/public/ddi/rankImages/` placeholder exists.
* Counts any DDI rank images already staged.

### 3.2 MOS cords and pins

Run `scripts/validate-mos.mjs`:

```bash
node scripts/validate-mos.mjs
```

It checks:

* `SHOULDER_CORD_MAP` resolves every configured code to a cord name.
* `NECK_PIN_MAP` resolves every configured code to a pin name.
* `MosGroup` membership is correct.
* Officer/enlisted MOS validation works.

### 3.3 Awards

Run `scripts/validate-awards.mjs`:

```bash
node scripts/validate-awards.mjs
```

It checks:

* Official award names resolve through `normalizeAwardName`.
* Valor-device names still resolve after stripping.
* Aliases map to canonical names.

Run `npm run sprites:generate` to verify the manifest and sprite sheets are consistent.

### 3.4 Badges and tabs

Run `node scripts/validate-uniform-assets.mjs` (Phase 5.8):

```bash
node scripts/validate-uniform-assets.mjs
```

It checks:

* `combatBadges/1.png` through `combatBadges/12.png` exist.
* `tabSpriteSheet.png` exists.
* `weaponQuals/` plates and root files are present.
* Ribbon and medal sprite sheets exist and are non-empty.
* All `mosConfig.mjs` cord/pin names have matching PNGs.

## 4. Smoke Test Checklist for Generated Uniforms

Use a known profile with a stable set of awards, ranks, MOS, and tabs. Render the uniform and visually inspect:

- [ ] Rank epaulette is correct for the `rankGrade`.
- [ ] Shoulder cord matches the MOS (or is absent if none).
- [ ] Neck pin matches the MOS (or is absent if none).
- [ ] Ribbons are in the correct order and no award is missing.
- [ ] Medal rack shows medals for awards with `medalPriority >= 2`.
- [ ] Oak leaf / star / numeral devices render in the correct count and position.
- [ ] Combat badge is appropriate for the MOS (CIB, EIB, aviation, medical, etc.).
- [ ] Weapon qualification bar is correct for the profile.
- [ ] Tab is visible if the member has a tab.
- [ ] Unit citation is present if earned.
- [ ] Name tag is rendered.
- [ ] No missing-image errors in the browser console.

## 5. Pre-Production Handoff

1. Confirm `ASSET_BASE_URL` points to the DDI staging host.
2. Run all validation scripts.
3. Generate sprite sheets from `uniform-uploads/`.
4. Render a sample uniform for one member of each rank and MOS.
5. Compare the output to the same profile rendered with 7CAV assets.
6. If identical, point `ASSET_BASE_URL` to the DDI production host and deploy.

## 6. Rollback

1. Revert `ASSET_BASE_URL` to `https://7cav.us` in the DDI API and BFF.
2. Restore `client/public/skunkworks/` sprite sheets from Git.
3. Revert `canvas.jsx` path changes if DDI paths were tested.

## 7. Summary

* All DDI asset testing is non-destructive: the legacy 7CAV files remain in `client/public/skunkworks/`.
* `ASSET_BASE_URL` and `client/public/ddi/` are the two levers for testing DDI assets.
* Validation scripts catch mismatches before any sprite rebuild or deployment.
* The smoke test checklist ensures the rendered uniform is visually correct end-to-end.