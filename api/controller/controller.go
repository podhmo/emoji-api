// Generated by swagger/tools/gen-stub --doc openapi.json --src ./api/oapigen --dst ./api/controller
package controller

// ControllerController :
type ControllerController struct {
	*EmojiController
}

// NewControllerController :
func NewControllerController() *ControllerController{
	return &ControllerController{
		EmojiController: NewEmojiController(),
	}
}
