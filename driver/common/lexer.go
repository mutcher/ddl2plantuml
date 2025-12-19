package common

type Lexer struct {
	index int
	ddl   []rune
}

func (l *Lexer) Reset(ddl *string) {
	l.ddl = []rune(*ddl)
	l.index = 0
}

func (l *Lexer) IsValid() bool {
	return l.index >= 0 && l.index < len(l.ddl)
}

func (l *Lexer) Current() rune {
	return l.ddl[l.index]
}

func (l *Lexer) Decr() {
	l.index--
}

func (l *Lexer) Inc() {
	l.index++
}

func (l *Lexer) Next() rune {
	l.Inc()
	return l.Current()
}

func (l *Lexer) PreviewNext() rune {
	l.Inc()
	result := l.Current()
	l.Decr()
	return result
}

func (l *Lexer) Prev() rune {
	l.Decr()
	return l.Current()
}
