package oracle

import (
	"fmt"
	"slices"

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

// CreateTableColumnsDefinitionSubState
const (
	COLUMN_NAME_STEP = iota
	COLUMN_TYPE_STEP
)

type CreateTableColumnsDefinitionState struct {
	Table     *common.Table
	TmpColumn common.Column
	Step      int
}

// GetTables implements MutableState.
func (s *CreateTableColumnsDefinitionState) GetTables() *common.Tables {
	panic("unimplemented")
}

func (s *CreateTableColumnsDefinitionState) InjectWord(word string) (MutableState, error) {
	if s.Step == COLUMN_TYPE_STEP && (isSqlEquals(word, ComaWord) || isSqlEquals(word, CloseBracketWord)) {
		if isSqlEquals(word, CloseBracketWord) {
			s.TmpColumn.Type += word
		}
		s.Table.Columns = append(s.Table.Columns, s.TmpColumn)
		return nil, nil
	}

	if s.Step == COLUMN_NAME_STEP {
		s.TmpColumn.Name = word
		s.Step = COLUMN_TYPE_STEP
		return s, nil
	}

	if s.Step == COLUMN_TYPE_STEP && !isSqlEquals(word, ComaWord) {
		s.TmpColumn.Type += word
		return s, nil
	}

	return s, nil
}

// CommentState
const (
	COMMENT_STATE_STEP_ON = iota
	COMMENT_STATE_STEP_COLUMN_OR_TABLE
	COMMENT_STATE_STEP_IS
	COMMENT_STATE_STEP_COLUMN_NAME
	COMMENT_STATE_STEP_TABLE_NAME
	COMMENT_STATE_STEP_COLUMN_COMMENT
	COMMENT_STATE_STEP_TABLE_COMMENT
)

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

type CommentState struct {
	Tables     *common.Tables
	EntityName string
	Step       int
}

// GetTables implements MutableState.
func (s *CommentState) GetTables() *common.Tables {
	return s.Tables
}

func ComentStateInit(tables *common.Tables) CommentState {
	return CommentState{Tables: tables, EntityName: "", Step: COMMENT_STATE_STEP_ON}
}

func (s *CommentState) InjectWord(word string) (MutableState, error) {
	if s.Step == COMMENT_STATE_STEP_TABLE_COMMENT {
		idx := slices.IndexFunc(*s.Tables, func(t common.Table) bool {
			return t.Name == s.EntityName
		})
		if idx == -1 {
			return s, TableNotFoundError(s.EntityName)
		}

		(*s.Tables)[idx].Comment = word
		return &NoneState{s.Tables}, nil
	}

	if s.Step == COMMENT_STATE_STEP_IS && isSqlEquals(word, IsWord) {
		s.Step = COMMENT_STATE_STEP_TABLE_COMMENT
	}

	if s.Step == COMMENT_STATE_STEP_TABLE_NAME {
		s.Step = COMMENT_STATE_STEP_IS
		s.EntityName = word
	}

	if s.Step == COMMENT_STATE_STEP_COLUMN_OR_TABLE && isSqlEquals(word, TableWord) {
		s.Step = COMMENT_STATE_STEP_TABLE_NAME
	}

	if s.Step == COMMENT_STATE_STEP_COLUMN_OR_TABLE && isSqlEquals(word, ColumnWord) {
		s.Step = COMMENT_STATE_STEP_COLUMN_NAME
	}

	if s.Step == COMMENT_STATE_STEP_ON && isSqlEquals(word, OnWord) {
		s.Step = COMMENT_STATE_STEP_COLUMN_OR_TABLE

	}

	return s, nil
}

// CreateTableState

type CreateTableState struct {
	Tables   *common.Tables
	TmpTable common.Table
	LastWord string

	SubState MutableState
}

// GetTables implements MutableState.
func (s *CreateTableState) GetTables() *common.Tables {
	return s.Tables
}

func CreateTableStateInit(tables *common.Tables) CreateTableState {
	return CreateTableState{Tables: tables, TmpTable: common.Table{}, LastWord: CreateWord, SubState: nil}
}

func (s *CreateTableState) InjectWord(word string) (MutableState, error) {
	if s.SubState != nil {
		subState, err := s.SubState.InjectWord(word)
		if err != nil {
			return nil, err
		}
		s.SubState = subState
	}

	// CREATE <...> TABLE
	if isSqlEquals(s.LastWord, CreateWord) && !isSqlEquals(word, TableWord) {
		return s, UnexpectedWordError{UnexpectedWord: word, PreviousWord: s.LastWord}
	}

	// CREATE TABLE <...>
	if isSqlEquals(s.LastWord, TableWord) {
		s.TmpTable.Name = word
	}

	// CREATE TABLE name (<...>)
	if s.LastWord == s.TmpTable.Name && isSqlEquals(word, OpenBracketWord) {
		s.SubState = &CreateTableColumnsDefinitionState{Table: &s.TmpTable, Step: COLUMN_NAME_STEP}
	}

	// CREATE TABLE (...)<;>
	if isSqlEquals(word, SemicolonWord) {
		newTables := append(*s.Tables, s.TmpTable)
		s.Tables = &newTables
		return &NoneState{&newTables}, nil
	}

	s.LastWord = word
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
