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

type SuggestOption struct {
	Limit   int
	Reverse bool
}

type pair struct {
	left  string
	right string
}

var pool = &sync.Pool{
	New: func() any {
		source := emoji.Map()
		r := make([]pair, len(source))
		i := 0
		for alias, char := range source {
			r[i] = pair{left: alias, right: char}
			i++
		}
		sort.SliceStable(r, func(i, j int) bool { return r[i].left < r[j].left })
		return r
	},
}

// Suggest returns the suggestions.
func Suggest(prefix string, option SuggestOption) []string {
	candidates := pool.Get().([]pair) // more cache?
	defer pool.Put(candidates)

	limit := option.Limit
	reversed := option.Reverse

	var r []string
	if limit != 0 {
		r = make([]string, 0, limit)
	}

	// O(N) (use trie?)
	if !reversed {
		for _, p := range candidates {
			if !strings.HasPrefix(p.left, prefix) {
				continue
			}

			r = append(r, p.right)
			if limit > 0 && len(r) >= limit {
				break
			}
		}
	} else {
		for i := len(candidates) - 1; i >= 0; i-- {
			p := candidates[i]
			if !strings.HasPrefix(p.left, prefix) {
				continue
			}

			r = append(r, p.right)
			if limit > 0 && len(r) >= limit {
				break
			}
		}
	}
	return r
}
