package controller

import (
	"context"

	oapigen "github.com/podhmo/emoji-api/api/oapigen"
	"github.com/podhmo/emoji-api/emojilib"
)

type EmojiController struct{}

func NewEmojiController() *EmojiController {
	return &EmojiController{}
}

// Suggest is endpoint of POST /emoji/suggest
// 先頭一致で対応する文字列を探す
//
// * body  :requestBody                         -- "need: var body oapigen.SuggestJSONBody; gctx.ShouldBindJSON(&body); "
func (c *EmojiController) Suggest(ctx context.Context, request oapigen.SuggestRequestObject) (response oapigen.SuggestResponseObject, err error) {
	prefix := request.Body.Prefix

	option := emojilib.SuggestOption{}
	if limit := request.Body.Limit; limit != nil {
		option.Limit = *limit
	}
	if sort := request.Body.Sort; sort == oapigen.SuggestJSONBodySortDesc {
		option.Reverse = true
	}

	suggestions := emojilib.Suggest(prefix, option)
	got := make([]oapigen.EmojiDefinition, len(suggestions))
	for i, x := range suggestions {
		got[i] = oapigen.EmojiDefinition{
			Alias: x.Alias,
			Char:  x.Char,
		}
	}
	response = oapigen.Suggest200JSONResponse(got)
	return
}

// Translate is endpoint of POST /emoji/translate
// :<alias>:のような表現を含んだ文字列をemojiを使った文字列に変換する
//
// * body  :requestBody                         -- "need: var body oapigen.TranslateJSONBody; gctx.ShouldBindJSON(&body); "
func (c *EmojiController) Translate(ctx context.Context, request oapigen.TranslateRequestObject) (response oapigen.TranslateResponseObject, err error) {
	text := request.Body.Text
	translated := emojilib.Translate(text)

	response = oapigen.Translate200JSONResponse(translated)
	return
}
