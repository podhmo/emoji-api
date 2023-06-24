package action

import "github.com/podhmo/emoji-api/seed/design"

var b = design.Builder

var (
	EmojiTranslate = b.Action("translate",
		b.Input(b.Body(b.Object(b.Field("text", b.String())))),
		b.Output(b.String()),
	).Doc(":<alias>:のような表現を含んだ文字列をemojiを使った文字列に変換する")

	EmojiSuggest = b.Action("suggest",
		b.Input(b.Body(
			b.Object(
				b.Field("prefix", b.String()),
				b.Field("sort", b.String().Enum([]string{"asc", "desc"}).Default("asc")),
				b.Field("limit", b.Int()).Required(false),
			)),
		),
		b.Output(b.Array(design.EmojiDefinition)),
	).Doc("先頭一致で対応する文字列を探す")
)
