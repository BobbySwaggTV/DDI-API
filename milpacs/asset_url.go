package milpacs

import (
	"strings"

	"github.com/spf13/viper"
)

const defaultAssetBaseURL = "https://7cav.us"

// assetBaseURL returns the configured ASSET_BASE_URL with any trailing slash
// removed. When the env var is unset, the legacy 7cav.us base is used so
// existing tests and unconfigured deployments continue to work.
func assetBaseURL() string {
	u := strings.TrimRight(viper.GetString("ASSET_BASE_URL"), "/")
	if u == "" {
		return defaultAssetBaseURL
	}
	return u
}
