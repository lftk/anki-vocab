package wordlist

import (
	"bufio"
	"iter"
	"os"
	"strings"
)

type Word struct {
	Text string
	Tags []string
}

type Deck struct {
	Name  string
	Words []*Word
}

func Load(path string) iter.Seq2[*Deck, error] {
	return func(yield func(*Deck, error) bool) {
		f, err := os.Open(path)
		if err != nil {
			yield(nil, err)
			return
		}
		defer f.Close()

		deck := new(Deck)
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)

			if line == "" || strings.HasPrefix(line, "//") {
				continue
			}

			if deckName, ok := strings.CutPrefix(line, "## "); ok {
				if deckName != deck.Name {
					if !yield(deck, nil) {
						return
					}
					deck = &Deck{Name: deckName}
				}
				continue
			}

			var tags []string
			word, after, ok := strings.Cut(line, "#")
			if ok {
				for tag := range strings.SplitSeq(after, "#") {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						tags = append(tags, tag)
					}
				}
			}

			word = strings.TrimSpace(word)
			if word != "" {
				deck.Words = append(deck.Words, &Word{
					Text: word,
					Tags: tags,
				})
			}
		}

		if !yield(deck, nil) {
			return
		}

		if err = scanner.Err(); err != nil {
			yield(nil, err)
			return
		}
	}
}
