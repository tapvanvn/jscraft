package tokenize

const (
	TokenUnknown = iota
)

type BaseToken struct {
	Type     int
	Content  string
	Children BaseTokenStream
}

func (token *BaseToken) GetType() int {
	return token.Type
}

func (token *BaseToken) GetContent() string {
	return token.Content
}

func (token *BaseToken) GetChildren() *BaseTokenStream {
	return &token.Children
}

func IndexOf(runes []rune, ch rune) int {
	tmpOffset := 0
	for {
		if tmpOffset == len(runes) {
			break
		}
		tmpRune := runes[tmpOffset]
		if tmpRune == ch {
			return tmpOffset
		}
		tmpOffset++
	}
	return -1
}
