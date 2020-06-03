package tokenize

const (
	TokenWord = iota
	TokenSpace
)

//WordTokenStream word
type WordTokenStream struct {
	Source string
	Tokens []BaseToken
	Offset int
	runes  []rune
}

func (stream *WordTokenStream) GetCurrentCharacter() rune {
	return stream.runes[stream.Offset]
}

func (stream *WordTokenStream) ReadCharacter() rune {
	if !stream.EOS() {
		var tmpOffset = stream.Offset
		stream.Offset++
		return stream.runes[tmpOffset]
	}
	return rune(0)
}

func (stream *WordTokenStream) NextIndexOf(ch rune) int {
	tmpOffset := stream.Offset + 1
	for {
		if tmpOffset == stream.Length() {
			break
		}
		tmpRune := stream.runes[tmpOffset]
		if tmpRune == ch {
			return tmpOffset
		}
		tmpOffset++
	}
	return -1
}

func (stream *WordTokenStream) GetToCharacter(toRune rune) string {
	var rs string = ""
	if !stream.EOS() {
		var pos int = stream.NextIndexOf(toRune)
		if pos >= stream.Offset {
			rs = string(stream.runes[stream.Offset:pos])
		}
	}
	return rs
}

func (stream *WordTokenStream) ReadToCharacter(toRune rune) string {
	var rs string = ""
	if !stream.EOS() {
		var pos int = stream.NextIndexOf(toRune)
		if pos >= stream.Offset {
			rs = string(stream.runes[stream.Offset:pos])
			stream.Offset = pos
		}
	}
	return rs
}

func (stream *WordTokenStream) ReadWhileCharacterIn(filter string) string {

	var runeFilter []rune = []rune(filter)
	var rsRune []rune
	for {
		if stream.EOS() {
			break
		}
		ch := stream.GetCurrentCharacter()
		var found bool = false
		for _, runeCh := range runeFilter {
			if runeCh == ch {
				found = true
				rsRune = append(rsRune, stream.ReadCharacter())
				break
			}
		}
		if !found {
			break
		}
	}
	return string(rsRune)
}

//Tokenize tokenize a string
func (stream *WordTokenStream) Tokenize(content string) {
	stream.Source = content
	stream.Offset = 0
	stream.runes = []rune(content)
}

//AddToken add token to stream
func (stream *WordTokenStream) AddToken(token BaseToken) {

}

//AddTokenByConntent AddTokenByConntent
func (stream *WordTokenStream) AddTokenByConntent(content []rune, tokenType int) {

}

//ReadToken read token
func (stream *WordTokenStream) ReadToken() BaseToken {
	return BaseToken{}
}

//ResetToBegin reset to begin
func (stream *WordTokenStream) ResetToBegin() {
	stream.Offset = 0
}

//EOS is end of stream
func (stream *WordTokenStream) EOS() bool {
	return stream.Offset >= len(stream.runes)
}

//Length get len of stream
func (stream *WordTokenStream) Length() int {
	return len(stream.runes)
}
