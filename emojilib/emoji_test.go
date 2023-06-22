package emojilib_test

import (
	"reflect"
	"testing"

	"github.com/podhmo/emoji-api/emojilib"
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
			if got := emojilib.Translate(tt.args.text); got != tt.want {
				t.Errorf("Translate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSuggest(t *testing.T) {
	type args struct {
		prefix string
		option emojilib.SuggestOption
	}
	tests := []struct {
		name string
		args args
		want []emojilib.Definition
	}{
		{name: "simple", args: args{prefix: ":diz", option: emojilib.SuggestOption{}},
			want: []emojilib.Definition{{":dizzy:", "💫"}, {":dizzy_face:", "😵"}},
		},
		{name: "simple-with-limit3", args: args{prefix: ":di", option: emojilib.SuggestOption{Limit: 3}},
			want: []emojilib.Definition{{":diamond_shape_with_a_dot_inside:", "💠"}, {":diamond_suit:", "♦️"}, {":diamond_with_a_dot:", "💠"}}},
		{name: "simple-with-reverse-limit3", args: args{prefix: ":di", option: emojilib.SuggestOption{Limit: 3, Reverse: true}},
			want: []emojilib.Definition{{":dizzy_face:", "😵"}, {":dizzy:", "💫"}, {":diya_lamp:", "🪔"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := emojilib.Suggest(tt.args.prefix, tt.args.option); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Suggest() = %v, want %v", got, tt.want)
			}
		})
	}
}
