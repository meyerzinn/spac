package fonts

import (
	"golang.org/x/image/font"
	"github.com/golang/freetype/truetype"
	"github.com/20zinnm/spac/client/data"
)

var (
	RobotoLight = loadTTF("../assets/fonts/Roboto-Light.ttf", 96)
)

func loadTTF(path string, size float64) (font.Face) {
	bytes := data.MustAsset(path)

	font, err := truetype.Parse(bytes)
	if err != nil {
		panic(err)
	}

	return truetype.NewFace(font, &truetype.Options{
		Size:              size,
		GlyphCacheEntries: 1,
	})
}
