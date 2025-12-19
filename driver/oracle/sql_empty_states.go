package oracle

import (
	"fmt"

	"github.com/icpd/ddl2plantuml/driver/common"
)

type MutableState interface {
	GetTables() *common.Tables
	InjectWord(word string) (MutableState, error)
}

type UnexpectedWordError struct {
	UnexpectedWord string
	PreviousWord   string
}

func (e UnexpectedWordError) Error() string {
	return fmt.Sprintf("Unexpected word in sequence %s prev word: %s", e.UnexpectedWord, e.PreviousWord)
}

// NullState

type NullState struct {
	Tables *common.Tables
}

// GetTables implements MutableState.
func (s *NullState) GetTables() *common.Tables {
	return s.Tables
}

func NullStateInit(tables *common.Tables) NullState {
	return NullState{Tables: tables}
}

func (s *NullState) InjectWord(word string) (MutableState, error) {
	if word == SemicolonWord {
		return &NoneState{s.Tables}, nil
	}

	return s, nil
}

// NoneState

type NoneState struct {
	Tables *common.Tables
}

// GetTables implements MutableState.
func (s *NoneState) GetTables() *common.Tables {
	return s.Tables
}

func (s *NoneState) InjectWord(word string) (MutableState, error) {
	if isSqlEquals(word, CreateWord) {
		newState := CreateTableStateInit(s.Tables)
		return &newState, nil
	}

	if isSqlEquals(word, CommentWord) {
		newState := ComentStateInit(s.Tables)
		return &newState, nil
	}

	return s, nil
}

// SkipState consumes tokens until the statement terminator (`;`).
type SkipState struct {
	Tables *common.Tables
}

func (s *SkipState) GetTables() *common.Tables {
	return s.Tables
}

func (s *SkipState) InjectWord(word string) (MutableState, error) {
	if isSqlEquals(word, SemicolonWord) {
		return &NoneState{s.Tables}, nil
	}
	return s, nil
}
