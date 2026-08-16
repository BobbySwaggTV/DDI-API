package milpacs

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestAssetBaseURL(t *testing.T) {
	prior := viper.GetString("ASSET_BASE_URL")
	defer viper.Set("ASSET_BASE_URL", prior)

	viper.Set("ASSET_BASE_URL", "")
	assert.Equal(t, "https://7cav.us", assetBaseURL())

	viper.Set("ASSET_BASE_URL", "https://assets.example.test/")
	assert.Equal(t, "https://assets.example.test", assetBaseURL())

	viper.Set("ASSET_BASE_URL", "https://assets.example.test")
	assert.Equal(t, "https://assets.example.test", assetBaseURL())
}

func TestRankImageURLUsesConfiguredBase(t *testing.T) {
	prior := viper.GetString("ASSET_BASE_URL")
	defer viper.Set("ASSET_BASE_URL", prior)
	viper.Set("ASSET_BASE_URL", "https://ddi.assets.test")

	r := Rank{RankId: 4, RankImage: 123}
	assert.Equal(t, "https://ddi.assets.test/data/roster_ranks/0/4.jpg?123", r.ImageURL())
}

func TestAwardImageURLUsesConfiguredBase(t *testing.T) {
	prior := viper.GetString("ASSET_BASE_URL")
	defer viper.Set("ASSET_BASE_URL", prior)
	viper.Set("ASSET_BASE_URL", "https://ddi.assets.test")

	a := Award{AwardId: 1, AwardImage: 5}
	assert.Equal(t, "https://ddi.assets.test/data/roster_awards/0/1.jpg?5", a.ImageURL())
}

func TestUniformUrlUsesConfiguredBase(t *testing.T) {
	prior := viper.GetString("ASSET_BASE_URL")
	defer viper.Set("ASSET_BASE_URL", prior)
	viper.Set("ASSET_BASE_URL", "https://ddi.assets.test")

	p := Profile{RelationId: 2}
	assert.Equal(t, "https://ddi.assets.test/data/roster_uniforms/0/2.jpg", p.UniformUrl())
}
