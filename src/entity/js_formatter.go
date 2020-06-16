package entity

import (
	"com.newcontinent-team.jscraft/tokenize"
	"com.newcontinent-team.jscraft/tokenize/js"
)

//JSFormatter ...
type JSFormatter struct {
	LastLineLength int
	LastLineLevel  int
	Content        string
}

type scoper struct {
	parent  *scoper
	content string
	level   int
}

func (s *scoper) feed(l *liner) {

	if len(l.content) > 0 {

		content := ""

		for i := 0; i < s.level; i++ {

			content += "\t"
		}
		s.content += "\n"

		content += l.content

		s.content += content
	}
	l.content = ""
}

func (s *scoper) finish() {

}

type liner struct {
	fortmatter *JSFormatter
	content    string

	last        tokenize.BaseToken
	lastMeaning tokenize.BaseToken

	s *scoper

	isGlue bool
}

func (l *liner) clear() {

	l.content = ""

	l.last.Type = 0

	l.lastMeaning.Type = 0
}

func (l *liner) feed(token *tokenize.BaseToken) {

	switch token.Type {

	case js.TokenJSWord:

		if l.lastMeaning.Type != 0 && l.lastMeaning.Type != js.TokenJSWord && l.last.Type != js.TokenJSPhraseBreak {

			l.content += " "
		}
		l.content += token.Content

		l.lastMeaning = *token

	case js.TokenJSPhraseStrongBreak:

		if token.Type == js.TokenJSPhraseStrongBreak {

			l.content += ";"
		}

	case js.TokenJSWordBreak:

		if l.last.Type == js.TokenJSWord {

			l.content += " "
		}

	case js.TokenJSGlueBegin:

		l.isGlue = true

	case js.TokenJSGlueEnd:

		l.isGlue = false

	case js.TokenJSOperator:

		l.content += token.Content

		l.lastMeaning = *token

	case js.TokenJSScopeBegin:

		l.s.feed(l)

		l.clear()

		newScope := &scoper{level: l.s.level + 1}

		newScope.parent = l.s

		l.s = newScope

		break

	case js.TokenJSScopeEnd:
		if l.s.parent != nil {
			//if len(l.s.content) > 0 {
			//	l.s.content += "\n"
			//}
			l.s.parent.content += l.s.content
			l.s = l.s.parent
		} else {
			//todo: next logic
		}
		break

	case js.TokenJSPhraseBreak:
		if l.isGlue {
			l.content += ";"
		} else {
			l.s.feed(l)
		}

	}

	l.last = *token
}

func (l *liner) start() {

	l.s = &scoper{}
}

func (l *liner) finish() {

	l.s.feed(l)

	for {
		if l.s.parent == nil {

			break
		}
		l.s.parent.content += l.s.content

		l.s = l.s.parent
	}

	l.fortmatter.Content = l.s.content
}

func (f *JSFormatter) formatStream(l *liner, stream *tokenize.BaseTokenStream) {

	stream.ResetToBegin()

	for {
		if stream.EOS() {

			break
		}

		token := stream.ReadToken()

		if token.Type != 0 || len(token.Content) > 0 {

			l.feed(token)

		} else if token.Children.Length() > 0 {

			f.formatStream(l, &token.Children)
		}
	}
}

//Format export stream
func (f *JSFormatter) Format(stream *tokenize.BaseTokenStream) {

	l := liner{fortmatter: f}

	l.start()

	f.formatStream(&l, stream)

	l.finish()
}
