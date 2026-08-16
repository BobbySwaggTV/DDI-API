# DDI MOS Migration Audit

> Phase 5.6: identify 7CAV-specific MOS content and its impact on the uniform builder.  
> No code, MOS values, rendering logic, API contracts, or assets were changed.

## 1. Current MOS Catalog

### 1.1 DDIApps MOS constants

The single source of truth for MOS codes in the uniform builder is `client/app/uniformbuilder/modules/constants/mos.js`.

| Category | Keys (UPPER_SNAKE) | Code | Notes |
| --- | --- | --- | --- |
| **Aviation** | `AVIATION_OFFICER` | `15A` | 7CAV S1 wiki-derived |
| | `ROTARY_WING_AVIATOR_WARRANT_OFFICER` | `153A` | |
| | `FIXED_WING_AVIATOR_WARRANT_OFFICER` | `155A` | |
| | `JET_AIRCRAFT_PILOT` | `155F` | |
| | `ENLISTED_ROTARY_CREWMAN` | `15T` | aircrew exception |
| **Infantry** | `INFANTRY_OFFICER` | `11A` | |
| | `INFANTRYMAN` | `11B` | |
| | `INDIRECT_FIRE_INFANTRYMAN` | `11C` | |
| **Medical** | `MEDICAL_OFFICER` | `67A` | |
| | `COMBAT_MEDIC` | `68W` | |
| **Armor / Cavalry** | `ARMOR_CAVALRY_OFFICER` | `19A` | |
| | `BRADLEY_CREWMEMBER` | `19C` | |
| | `CAVALRY_SCOUT` | `19D` | |
| | `ARMOR_CREWMAN` | `19K` | |
| **Engineer** | `COMBAT_ENGINEER_OFFICER` | `12A` | |
| | `COMBAT_ENGINEER` | `12B` | |
| **Field Artillery** | `FIELD_ARTILLERY_OFFICER` | `13A` | |
| | `CANNON_CREWMEMBER` | `13B` | |
| **Command** | `COMMAND_SERGEANT_MAJOR` | `00Z` | |
| | `OFFICER_GENERALIST` | `01A` | |
| **Staff — S1** | `S1_OFFICER` | `42B` | |
| | `S1_ENLISTED` | `42A` | |
| **Staff — S2** | `S2_OFFICER` | `35A` | |
| | `S2_NCO` | `35B` | not on 7CAV wiki, kept for roster compatibility |
| | `S2_ENLISTED` | `35F` | |
| **Staff — S3** | `S3_OFFICER` | `57A` | |
| | `S3_ENLISTED` | `57B` | |
| **Staff — S5** | `S5_OFFICER` | `46A` | |
| | `S5_ENLISTED` | `46S` | |
| **Staff — S6** | `S6_OFFICER` | `25A` | |
| | `S6_ENLISTED` | `25U` | |
| | `REGIMENTAL_TECHNICAL_AIDE` | `255N` | |
| **Staff — S7** | `S7_OFFICER` | `47A` | |
| **Military Police** | `MP_OFFICER` | `31A` | |
| | `MP_ENLISTED` | `31B` | |
| **JAG** | `JAG_OFFICER` | `27A` | |
| | `JAG_ENLISTED` | `27D` | |
| **RRD** | `RRD_OFFICER` | `79A` | |
| | `RRD_ENLISTED` | `79R` | |
| **RTC** | `RTC_OFFICER` | `79Z` | |
| | `RTC_ENLISTED` | `79X` | |
| **WAG** | `WAG_OFFICER` | `26Z` | |
| | `WAG_ENLISTED` | `26B` | |
| **Other** | `ODS_OFFICER` | `47Q` | |
| | `NCOA_OFFICER` | `47C` | |
| | `NCOA_ENLISTED` | `47T` | |
| | `FCC_ANALYST` | `49A` | |
| | `RDC_OFFICER` | `50A` | |
| | `DEVCOM_LEAD` | `51A` | |
| | `DEVCOM_SUPPORT_COORDINATOR` | `51S` | |

<ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/constants/mos.js" />

### 1.2 Semantic MOS groups

`mos.js` also exports `MosGroup`, which drives badge logic in `AwardClasses.jsx`:

| Group | Members | Effect on badge display |
| --- | --- | --- |
| `AVIATION` | `153A, 155A, 15A, 15T, 155F` | Marks the user as aviation for badge selection |
| `MEDICAL` | `68W, 67A` | Marks the user as medical for badge selection |
| `AIRCREW` | `15T` | Caps `maxAllowed` at `8` (aircrew-specific badge behavior) |

<ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/AwardClasses.jsx" />

### 1.3 DDI API MOS data

The DDI API does not define or validate MOS codes. It passes through whatever is stored in the `Mos` custom field on `xf_nf_rosters_user`.

| Layer | Source | Behavior |
| --- | --- | --- |
| Database custom field | `xf_nf_rosters_user` | Free-text `Mos` stored in a JSON custom-field blob. |
| GORM custom field | `milpacs/profiles.go` `CustomFields` | `Mos` is deserialized from the custom-field JSON. |
| API wire field | `types/profiles.go` `Profile.Mos` | plain string, no enum or validation. |
| Datastore | `datastores/mysql.go` `collectProfile` | `Mos: profile.UnmarshalCustomFields().Mos` |

<ref_file file="/home/bobbyswagg/research/DDI-API/types/profiles.go" />
<ref_file file="/home/bobbyswagg/research/DDI-API/datastores/mysql.go" />

## 2. MOS Mappings

### 2.1 Shoulder cord mappings

`client/app/uniformbuilder/modules/GetUserInfo.jsx` `setShoulderCord()` returns one of:

| Cord PNG | MOS codes that produce it |
| --- | --- |
| `Medical.png` | `68W`, `67A` |
| `Infantry.png` | `11B`, `11A`, `11C` |
| `Artillery.png` | `13A`, `13B` |
| `Engineer.png` | `12A`, `12B` |
| `Aide.png` | `01A` |
| `MP.png` | `31A`, `31B` |
| `Armor.png` | `19K`, `19A`, `19C`, `19D` |
| none (`false`) | any other code, or if `mos` is missing/unknown |

The canvas then loads `client/public/skunkworks/uniformCords/{shoulderCord}.png`.

<ref_file file="/home/bobbyswagg/research/CavApps/client/app/uniformbuilder/modules/GetUserInfo.jsx" />

### 2.2 Neck pin mappings

`client/app/uniformbuilder/modules/GetUserInfo.jsx` `setNeckPins()` returns one of:

| Pin PNG | MOS codes that produce it |
| --- | --- |
| `AviationOfficer` | `153A, 155A, 15A` |
| `AviationNCO` | `15T, 155F` |
| `ArtilleryOfficer` | `13A` |
| `ArtilleryNCO` | `13B` |
| `MedicalOfficer` | `67A` |
| `MedicalNCO` | `68W` |
| `EngineerOfficer` | `12A` |
| `EngineerNCO` | `12B` |
| `Aide` | `01A` |
| `CSM` | `00Z` |
| `IMOOfficer` | `255N, 25A` |
| `IMONCO` | `25U` |
| `S1S3S5` | `42B, 57A, 46A` |
| `S2NCO` | `35B` |
| `S2Officer` | `35A` |
| `MPOfficer` | `31A` |
| `MPNCO` | `31B` |
| `ArmorOfficer` | `19A` |
| `ArmorNCO` | `19K, 19C, 19D` |
| `JAGOfficer` | `27A` |
| `JAGNCO` | `27D` |
| `RDCOfficer` | `51A, 50A` |
| `InfantryOfficer` | `11A, 13A, 47A, 79A, 79Z, 26Z, 47Q, 47C` |
| `InfantryNCO` | `11B, 11C, 42A, 57B, 46S, 79R, 79X, 51S, 49A, 26B, 47T` |
| none (`false`) | any other code, or missing `mos` |

The canvas then loads `client/public/skunkworks/uniformLapelPins/{neckPins}.png`.

### 2.3 MOS validation

`checkMos()` in `GetUserInfo.jsx` uses a regex to ensure officer-coded MOS (`[0-9]+[A,B,Z,Q,C,N]`) are not assigned to enlisted/WO members and vice versa. The regex contains 7CAV-specific exceptions (`00Z, 11C, 42A, 49A, 14B, 12B, 35B, 31B, 11B, 57B, 26B, 13B, 19C`).

## 3. Which MOS Values Are Inherited from 7CAV

Everything in `mos.js` is inherited from the 7CAV S1 wiki:

* The UPPER_SNAKE key names (e.g. `INFANTRYMAN`).
* The code strings (e.g. `11B`).
* The staff-section naming (`S1`, `S2`, `S3`, `S5`, `S6`, `S7`).
* The special DMI cells (`RRD`, `RTC`, `WAG`, `ODS`, `NCOA`, `FCC`, `RDC`, `DEVCOM`).
* The wiki URL in the file header.

The DDI API is MOS-agnostic; it only echoes the string.

## 4. Which Mappings Can Remain

| Mapping | Why it can stay | Caveat |
| --- | --- | --- |
| The generic code strings `11B`, `68W`, `13A`, etc. | They are U.S. Army MOS codes, not 7CAV-specific. | DDI may drop some or add custom codes. |
| The `MosGroup.AVIATION` / `MEDICAL` / `AIRCREW` groups | They are semantic groupings, not brand names. | Membership must match the new DDI MOS values if codes change. |
| The shoulder-cord / neck-pin image filenames | They are abstract (`Infantry.png`, `MedicalNCO.png`). | No rebrand needed unless the art changes. |
| The `Profile.Mos` wire shape | It is a plain string, organization-neutral. | None. |

## 5. Which Mappings Need DDI Ownership

| Content | Why it must change | Recommended action |
| --- | --- | --- |
| `mos.js` catalog keys and comments | Names like `7th Cavalry wiki` and `S1/S2/S3` are 7CAV organizational language. | Build a DDI MOS catalog with DDI section names. |
| Staff-section codes (`42B`, `35A`, `57A`, etc.) | These are 7CAV staff section mappings (S1, S2, S3, S5, S6, S7). | Re-map to DDI staff sections or keep the same codes but update comments. |
| `RRD`, `RTC`, `WAG`, `ODS`, `NCOA`, `FCC`, `RDC`, `DEVCOM` codes | These are 7CAV-specific commands/schools. | Replace with DDI command/school codes and neck-pin mappings. |
| The 7CAV wiki URL | It is a 7CAV link. | Replace with DDI documentation or remove. |
| `MosGroup` membership | It is hardcoded to 7CAV MOS values. | Update if DDI adds/removes MOS codes. |
| `setShoulderCord()` and `setNeckPins()` switch statements | They are hardcoded to the 7CAV catalog. | Update to the new DDI catalog. |
| `checkMos()` regex exceptions | The exception list is 7CAV-specific. | Rebuild with DDI MOS regex and exceptions. |

## 6. Risks

### 6.1 Shoulder cords

* **Risk:** A DDI MOS string does not match any `case` in `setShoulderCord()`.
* **Impact:** The canvas skips the shoulder cord. The uniform looks incomplete.
* **Mitigation:** Ensure every DDI MOS that should have a cord is present in `setShoulderCord()`. Keep the existing `default: false` so unknown codes render without a cord rather than crashing.

### 6.2 Neck pins

* **Risk:** A DDI MOS string does not match any `case` in `setNeckPins()`.
* **Impact:** The canvas skips the lapel pin.
* **Mitigation:** Add the new DDI MOS codes to `setNeckPins()` or map them to a generic `InfantryNCO`/`InfantryOfficer` fallback. Keep the `default: false`.

### 6.3 Uniform rendering

* **Risk:** `MosGroup.AVIATION` / `MEDICAL` / `AIRCREW` no longer match the DDI MOS codes.
* **Impact:** Aviation and medical badges are not rendered correctly; 15T aircrew may not receive the correct badge cap.
* **Mitigation:** Update `MosGroup` membership to the new DDI aviation and medical codes before changing the catalog keys.

### 6.4 MOS validation

* **Risk:** `checkMos()` flags valid DDI officer codes as enlisted because the regex exceptions are 7CAV-specific.
* **Impact:** The builder shows `MOS ... is an Enlisted/NCO MOS, despite user being of Officer/WO rank`.
* **Mitigation:** Rebuild `checkMos()` with the DDI officer-code pattern and exception list.

### 6.5 API contract

* **Risk:** None for the current API. It already passes `mos` as a string.
* **Caveat:** If DDI changes the custom-field key from `mos` to something else, both `milpacs/profiles.go` and the API response will break. Keep the `mos` JSON key.

## 7. Summary

* The DDI API has no MOS content of its own; it echoes the `mos` custom-field string.
* All MOS logic, catalog, and mappings live in the DDIApps frontend (`mos.js` and `GetUserInfo.jsx`).
* The U.S. Army code strings (`11B`, `68W`, etc.) are the most portable parts and can remain.
* The 7CAV-specific staff sections, commands, schools, wiki URL, and the switch statements in `GetUserInfo.jsx` need DDI ownership.
* The safest migration is:
  1. Keep the `mos` wire shape unchanged.
  2. Build the DDI MOS catalog.
  3. Update `MosGroup` membership.
  4. Update `setShoulderCord()` and `setNeckPins()`.
  5. Update `checkMos()`.
  6. Replace the `uniformCords/` and `uniformLapelPins/` artwork if DDI wants different imagery.