package emoji_test

import (
	"reflect"
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

func TestSuggest(t *testing.T) {
	type args struct {
		prefix string
		option emoji.SuggestOption
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "simple", args: args{prefix: ":di", option: emoji.SuggestOption{}}, want: []string{"💠", "♦️", "💠", "♦️", "🇩🇬", "🔅", "🎯", "😞", "😞", "😥", "🥸", "➗", "🤿", "🪔", "💫", "😵"}},
		{name: "simple-with-limit3", args: args{prefix: ":di", option: emoji.SuggestOption{Limit: 3}}, want: []string{"💠", "♦️", "💠"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := emoji.Suggest(tt.args.prefix, tt.args.option); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Suggest() = %v, want %v", got, tt.want)
			}
		})
	}
}
