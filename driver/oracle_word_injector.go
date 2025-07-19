package driver

import (
	"fmt"
	"slices"
)

const (
	CreateWord       = "CREATE"
	CommentWord      = "COMMENT"
	ColumnWord       = "COLUMN"
	TableWord        = "TABLE"
	OnWord           = "ON"
	OpenBracketWord  = "("
	CloseBracketWord = ")"
	ComaWord         = ","
	SemicolonWord    = ";"
)

type UnexpectedWordError struct {
	UnexpectedWord string
	PreviousWord   string
}

func (e UnexpectedWordError) Error() string {
	return fmt.Sprintf("Unexpected word in sequence %s prev word: %s", e.UnexpectedWord, e.PreviousWord)
}

type MutableState interface {
	GetTables() *Tables
	InjectWord(word string) (MutableState, error)
}

// CreateTableColumnsDefinitionSubState
const (
	COLUMN_NAME_STEP = iota
	COLUMN_TYPE_STEP
)

type CreateTableColumnsDefinitionSubState struct {
	Table     *Table
	TmpColumn Column
	Step      int
}

// GetTables implements MutableState.
func (s *CreateTableColumnsDefinitionSubState) GetTables() *Tables {
	panic("unimplemented")
}

func (s *CreateTableColumnsDefinitionSubState) InjectWord(word string) (MutableState, error) {
	if s.Step == COLUMN_TYPE_STEP && !isSqlEquals(word, ComaWord) {
		s.TmpColumn.Type += word
	}

	if s.Step == COLUMN_NAME_STEP {
		s.TmpColumn.Name = word
		s.Step = COLUMN_TYPE_STEP
	}

	if s.Step == COLUMN_TYPE_STEP && isSqlEquals(word, ComaWord) {
		s.Table.Columns = append(s.Table.Columns, s.TmpColumn)
		return nil, nil
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
	Tables *Tables
}

// GetTables implements MutableState.
func (s *NullState) GetTables() *Tables {
	return s.Tables
}

func NullStateInit(tables *Tables) NullState {
	return NullState{Tables: tables}
}

func (s *NullState) InjectWord(word string) (MutableState, error) {
	if word == SemicolonWord {
		return &NoneState{s.Tables}, nil
	}

	return s, nil
}

type CommentState struct {
	Tables     *Tables
	EntityName string
	Step       int
}

// GetTables implements MutableState.
func (s *CommentState) GetTables() *Tables {
	return s.Tables
}

func ComentStateInit(tables *Tables) CommentState {
	return CommentState{Tables: tables, EntityName: "", Step: COMMENT_STATE_STEP_ON}
}

func (s *CommentState) InjectWord(word string) (MutableState, error) {
	if s.Step == COMMENT_STATE_STEP_TABLE_COMMENT {
		idx := slices.IndexFunc(*s.Tables, func(t Table) bool {
			return t.Name == s.EntityName
		})
		if idx == -1 {
			return s, TableNotFoundError(s.EntityName)
		}

		(*s.Tables)[idx].Comment = word
		return &NoneState{s.Tables}, nil
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
	Tables   *Tables
	TmpTable Table
	LastWord string

	SubState MutableState
}

// GetTables implements MutableState.
func (s *CreateTableState) GetTables() *Tables {
	return s.Tables
}

func CreateTableStateInit(tables *Tables) CreateTableState {
	return CreateTableState{Tables: tables, TmpTable: Table{}, LastWord: CreateWord, SubState: nil}
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
		s.SubState = &CreateTableColumnsDefinitionSubState{Table: &s.TmpTable, Step: COLUMN_NAME_STEP}
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
	Tables *Tables
}

// GetTables implements MutableState.
func (s *NoneState) GetTables() *Tables {
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

// SqlStateHandlerImpl

type SqlStateHandlerImpl struct {
	Tables       *Tables
	CurrentState MutableState
}

// GetTables implements SqlStateHandler.
func (s *SqlStateHandlerImpl) GetTables() *Tables {
	return s.CurrentState.GetTables()
}

func (s *SqlStateHandlerImpl) Reset() {
	s.Tables = &Tables{}
	s.CurrentState = &NoneState{s.Tables}
}

func (s *SqlStateHandlerImpl) InjectWord(word string) error {
	state, err := s.CurrentState.InjectWord(word)
	if err != nil {
		return err
	}

	s.CurrentState = state
	return nil
}
