package emoji

import (
	"sort"
	"strings"
	"sync"

	"github.com/enescakir/emoji"
)

// Trnaslate translates `:<emoji>:` to actual emoji unicode.
func Translate(text string) string {
	return emoji.Parse(text)
}

// Suggest returns the suggestions.
func Suggest(prefix string, option SuggestOption) []Definition {
	candidates := pool.Get().([]Definition) // more cache?
	defer pool.Put(candidates)

	limit := option.Limit
	reversed := option.Reverse

	var r []Definition
	if limit != 0 {
		r = make([]Definition, 0, limit)
	}

	// O(N) (use trie?)
	if !reversed {
		for _, p := range candidates {
			if !strings.HasPrefix(p.Alias, prefix) {
				continue
			}

			r = append(r, p)
			if limit > 0 && len(r) >= limit {
				break
			}
		}
	} else {
		for i := len(candidates) - 1; i >= 0; i-- {
			p := candidates[i]
			if !strings.HasPrefix(p.Alias, prefix) {
				continue
			}

			r = append(r, p)
			if limit > 0 && len(r) >= limit {
				break
			}
		}
	}
	return r
}

type SuggestOption struct {
	Limit   int
	Reverse bool
}

type Definition struct {
	Alias string
	Char  string
}

var pool = &sync.Pool{
	New: func() any {
		source := emoji.Map()
		r := make([]Definition, len(source))
		i := 0
		for alias, char := range source {
			r[i] = Definition{Alias: alias, Char: char}
			i++
		}
		sort.SliceStable(r, func(i, j int) bool { return r[i].Alias < r[j].Alias })
		return r
	},
}
