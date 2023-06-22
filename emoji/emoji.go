package emoji

import "github.com/enescakir/emoji"


// Trnaslate translates `:<emoji>:` to actual emoji unicode.
func Translate(text string) string {
	return emoji.Parse(text)
}
