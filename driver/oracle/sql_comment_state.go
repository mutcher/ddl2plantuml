package oracle

import (
	"slices"

	"github.com/icpd/ddl2plantuml/driver/common"
)

const (
	COMMENT_STATE_STEP_ON = iota
	COMMENT_STATE_STEP_COLUMN_OR_TABLE
	COMMENT_STATE_STEP_IS
	COMMENT_STATE_STEP_COLUMN_NAME
	COMMENT_STATE_STEP_TABLE_NAME
	COMMENT_STATE_STEP_COLUMN_COMMENT
	COMMENT_STATE_STEP_TABLE_COMMENT
)

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
