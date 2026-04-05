package source

import "regexp"

var (
	xmltvChannelBlock   = regexp.MustCompile(`(?s)<channel\b[^>]*>.*?</channel>`)
	xmltvChannelIconTag = regexp.MustCompile(`(?s)<icon\b[^>]*/>|<icon\b[^>]*>.*?</icon>`)
)

// stripChannelIconsFromXMLTV removes <icon> elements inside <channel>...</channel> only.
// Upstream Jellyfin overwrites M3U tvg-logo with XMLTV <channel><icon> when it thinks it matched
// the EPG row; imperfect matching yields wrong channel art. Programme-level <icon> is left intact.
func stripChannelIconsFromXMLTV(data []byte) []byte {
	return xmltvChannelBlock.ReplaceAllFunc(data, func(block []byte) []byte {
		return xmltvChannelIconTag.ReplaceAll(block, nil)
	})
}
