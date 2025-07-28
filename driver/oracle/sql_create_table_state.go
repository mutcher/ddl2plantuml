package oracle

import "github.com/icpd/ddl2plantuml/driver/common"

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
