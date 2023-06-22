package emoji_test

import (
	"testing"

	"github.com/podhmo/emoji-api/emoji"
)

func TestTranslate(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "simple", args: args{text: "(o_0) :dizzy:"}, want: "(o_0) 💫"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := emoji.Translate(tt.args.text); got != tt.want {
				t.Errorf("Translate() = %v, want %v", got, tt.want)
			}
		})
	}
}
