package oracle

import (
	"strings"

	"github.com/icpd/ddl2plantuml/driver/common"
)

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
	// This substate doesn't own the full tables slice; return an empty placeholder.
	empty := common.Tables{}
	return &empty
}

func (s *CreateTableColumnsDefinitionState) InjectWord(word string) (MutableState, error) {
	if s.Step == COLUMN_TYPE_STEP {
		// compute current parenthesis depth from the type built so far
		preDepth := strings.Count(s.TmpColumn.Type, "(") - strings.Count(s.TmpColumn.Type, ")")
		// handle parentheses and separators
		if isSqlEquals(word, OpenBracketWord) {
			s.TmpColumn.Type += word
			return s, nil
		}

		if isSqlEquals(word, CloseBracketWord) {
			if preDepth > 0 {
				s.TmpColumn.Type += word
				return s, nil
			}
			// closing the column list: finalize column and end substate
			s.Table.Columns = append(s.Table.Columns, s.TmpColumn)
			return nil, nil
		}

		if isSqlEquals(word, ComaWord) {
			if preDepth > 0 {
				// comma inside type args
				s.TmpColumn.Type += word
				return s, nil
			}
			// top-level comma: finalize column and prepare for next
			s.Table.Columns = append(s.Table.Columns, s.TmpColumn)
			s.Step = COLUMN_NAME_STEP
			return s, nil
		}

		// general token -> append with sensible spacing (no space after '(')
		if s.TmpColumn.Type == "" {
			s.TmpColumn.Type = word
		} else if strings.HasSuffix(s.TmpColumn.Type, "(") {
			s.TmpColumn.Type += word
		} else {
			s.TmpColumn.Type += word
		}

		return s, nil
	}

	if s.Step == COLUMN_NAME_STEP {
		s.TmpColumn.Name = word
		s.Step = COLUMN_TYPE_STEP
		return s, nil
	}

	return s, nil
}

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

	// CREATE <...> TABLE or other CREATE statements
	if isSqlEquals(s.LastWord, CreateWord) && !isSqlEquals(word, TableWord) {
		// If it's a CREATE INDEX (or other unsupported CREATE), skip until semicolon
		if isSqlEquals(word, IndexWord) {
			return &SkipState{Tables: s.Tables}, nil
		}
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
