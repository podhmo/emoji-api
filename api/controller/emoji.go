package controller

import (
	"context"
	oapigen "github.com/podhmo/emoji-api/api/oapigen"
)

type EmojiController struct { }

func NewEmojiController() *EmojiController {
	return &EmojiController{}
}


// Suggest is endpoint of POST /emoji/suggest
// 先頭一致で対応する文字列を探す
// 
// * body  :requestBody                         -- "need: var body oapigen.SuggestJSONBody; gctx.ShouldBindJSON(&body); "
func (c *EmojiController) Suggest(ctx context.Context, request oapigen.SuggestRequestObject) (output oapigen.SuggestResponseObject, err error) {
	return
}

// Translate is endpoint of POST /emoji/translate
// :<alias>:のような表現を含んだ文字列をemojiを使った文字列に変換する
// 
// * body  :requestBody                         -- "need: var body oapigen.TranslateJSONBody; gctx.ShouldBindJSON(&body); "
func (c *EmojiController) Translate(ctx context.Context, request oapigen.TranslateRequestObject) (output oapigen.TranslateResponseObject, err error) {
	return
}
