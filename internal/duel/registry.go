package duel

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

const LegacyRosterID = "legacy-20-v1"
const RegistryVersion = "duel-heroes-127-v1"
const RegistrySHA256 = "5bca2bf8c43972583d1d58c876f5dcde039cabb0e662ac9536b0e7a53037f138"

// The catalog is an identity map, not a declaration of playable heroes.
//
//go:embed registry/id-mapping.json
var registryJSON []byte

// New gameplay rosters require frontend acceptance and explicit build bindings.
// Never regenerate this list from the catalog or accept it from an HTTP client.
//
//go:embed registry/gameplay-rosters.json
var gameplayJSON []byte

//go:embed registry/protocol-features.json
var protocolJSON []byte

type HeroIdentity struct {
	HeroID      int    `json:"registryNumericId"`
	InternalID  string `json:"internalHeroId"`
	ValveID     int    `json:"valveHeroId"`
	ValveKey    string `json:"valveHeroKey"`
	LegacyIndex *int   `json:"legacyIndex"`
}

type GameplayRoster struct {
	ID              string   `json:"rosterId"`
	RegistryVersion string   `json:"registryVersion"`
	HeroIDs         []int    `json:"heroIds"`
	GameVersions    []string `json:"gameVersions"`
}

type heroRegistry struct {
	selectionBuilds map[string]bool
	heroes          []HeroIdentity
	byID            map[int]HeroIdentity
	rosters         []GameplayRoster
	byRoster        map[string]GameplayRoster
	byBuild         map[string]string
}

func loadRegistry(catalog, gameplay []byte) (*heroRegistry, error) {
	var manifest struct {
		Version string         `json:"registryVersion"`
		SHA     string         `json:"registrySha256"`
		Heroes  []HeroIdentity `json:"heroes"`
	}
	var raw struct {
		Heroes []map[string]any `json:"heroes"`
	}
	if json.Unmarshal(catalog, &manifest) != nil || json.Unmarshal(catalog, &raw) != nil {
		return nil, fmt.Errorf("invalid hero catalog")
	}
	// Go's map marshaler sorts keys, matching the frozen ASCII manifest's canonical form.
	if manifest.Version != RegistryVersion || manifest.SHA != RegistrySHA256 || digest(raw.Heroes) != RegistrySHA256 {
		return nil, fmt.Errorf("hero catalog hash mismatch")
	}
	r := &heroRegistry{heroes: manifest.Heroes, byID: map[int]HeroIdentity{}, byRoster: map[string]GameplayRoster{}, byBuild: map[string]string{}}
	valves, names := map[int]bool{}, map[string]bool{}
	for _, h := range manifest.Heroes {
		if _, exists := r.byID[h.HeroID]; exists || h.HeroID < 0 || h.ValveID <= 0 || h.InternalID == "" || valves[h.ValveID] || names[h.InternalID] {
			return nil, fmt.Errorf("duplicate or invalid hero identity")
		}
		r.byID[h.HeroID], valves[h.ValveID], names[h.InternalID] = h, true, true
	}
	if json.Unmarshal(gameplay, &r.rosters) != nil {
		return nil, fmt.Errorf("invalid gameplay rosters")
	}
	for _, roster := range r.rosters {
		if !versionPattern.MatchString(roster.ID) || roster.ID == RegistryVersion || roster.RegistryVersion != RegistryVersion || len(roster.HeroIDs) == 0 {
			return nil, fmt.Errorf("invalid gameplay roster")
		}
		if _, exists := r.byRoster[roster.ID]; exists {
			return nil, fmt.Errorf("duplicate gameplay roster")
		}
		seen := map[int]bool{}
		for _, id := range roster.HeroIDs {
			if _, ok := r.byID[id]; !ok || seen[id] {
				return nil, fmt.Errorf("unknown or duplicate gameplay hero")
			}
			seen[id] = true
		}
		if roster.ID == LegacyRosterID {
			if len(seen) != 20 || len(roster.GameVersions) != 0 {
				return nil, fmt.Errorf("legacy roster is immutable")
			}
			for id := 0; id < 20; id++ {
				if !seen[id] {
					return nil, fmt.Errorf("legacy roster is immutable")
				}
			}
		} else if len(roster.GameVersions) == 0 {
			return nil, fmt.Errorf("new gameplay roster requires approved game versions")
		}
		for _, build := range roster.GameVersions {
			if !versionPattern.MatchString(build) || r.byBuild[build] != "" {
				return nil, fmt.Errorf("invalid or duplicate game version binding")
			}
			r.byBuild[build] = roster.ID
		}
		r.byRoster[roster.ID] = roster
	}
	if _, ok := r.byRoster[LegacyRosterID]; !ok {
		return nil, fmt.Errorf("legacy roster missing")
	}
	return r, nil
}

func mustRegistry() *heroRegistry {
	r, err := loadRegistry(registryJSON, gameplayJSON)
	if err != nil {
		panic(err)
	}
	if err = r.loadFeatures(protocolJSON); err != nil {
		panic(err)
	}
	return r
}

// Empty roster is the old wire format. Do not normalize the request itself:
// its original serialization is used by already-persisted idempotency digests.
func (r *heroRegistry) resolve(id, version string, heroes ...int) (GameplayRoster, error) {
	if id == "" {
		id = LegacyRosterID
	}
	roster, ok := r.byRoster[id]
	if !ok {
		return roster, bad("unknown or inactive gameplay roster")
	}
	bound := r.byBuild[version]
	if (id != LegacyRosterID && bound != id) || (id == LegacyRosterID && bound != "") {
		return roster, conflict("game version and roster mismatch")
	}
	for _, hero := range heroes {
		found := false
		for _, member := range roster.HeroIDs {
			if hero == member {
				found = true
				break
			}
		}
		if !found {
			return roster, bad("hero is not playable in this roster")
		}
	}
	return roster, nil
}

func effectiveRoster(id string) string {
	if id == "" {
		return LegacyRosterID
	}
	return id
}

func effectiveRegistry(version string) string {
	if version == "" {
		return LegacyRosterID
	} // historical rows predate the catalog
	return version
}

// Explicit exact-build opt-in; an empty default never upgrades legacy rooms.
func (r *heroRegistry) loadFeatures(b []byte) error {
	var features struct {
		RoomSelectionVersions []string `json:"roomSelectionVersions"`
	}
	if json.Unmarshal(b, &features) != nil {
		return fmt.Errorf("invalid protocol features")
	}
	builds := map[string]bool{}
	for _, version := range features.RoomSelectionVersions {
		if !versionPattern.MatchString(version) || builds[version] || r.byBuild[version] == "" {
			return fmt.Errorf("selection requires a unique approved roster build")
		}
		roster := r.byRoster[r.byBuild[version]]
		found := false
		for _, hero := range roster.HeroIDs {
			if hero == 1 {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("selection roster requires default hero 1")
		}
		builds[version] = true
	}
	r.selectionBuilds = builds
	return nil
}
func (r *heroRegistry) selectionEnabled(version string) bool { return r.selectionBuilds[version] }
