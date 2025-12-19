package oracle

import (
	"slices"
	"strings"

	"github.com/icpd/ddl2plantuml/driver/common"
)

const (
	COMMENT_STATE_STEP_ON = iota
	COMMENT_STATE_STEP_COLUMN_OR_TABLE
	COMMENT_STATE_STEP_IS
	COMMENT_STATE_STEP_ENTITY_NAME
	COMMENT_STATE_STEP_COLUMN_COMMENT
	COMMENT_STATE_STEP_COMMENT_VALUE
)

type CommentState struct {
	Tables     *common.Tables
	EntityName string
	Step       int
	IsColumn   bool
}

// GetTables implements MutableState.
func (s *CommentState) GetTables() *common.Tables {
	return s.Tables
}

func ComentStateInit(tables *common.Tables) CommentState {
	return CommentState{Tables: tables, EntityName: "", Step: COMMENT_STATE_STEP_ON, IsColumn: false}
}

func (s *CommentState) InjectWord(word string) (MutableState, error) {
	if s.Step == COMMENT_STATE_STEP_COMMENT_VALUE {
		if s.IsColumn {
			entities := strings.Split(s.EntityName, ".")
			if len(entities) != 2 {
				return s, OracleSQLSyntaxError("Invalid column name format, expected TABLE.COLUMN")
			}
			tableName := entities[0]
			columnName := entities[1]
			idx := slices.IndexFunc(*s.Tables, func(t common.Table) bool {
				return t.Name == tableName
			})

			if idx == -1 {
				return s, TableNotFoundError(tableName)
			}

			table := &(*s.Tables)[idx]
			colIdx := slices.IndexFunc(table.Columns, func(c common.Column) bool {
				return c.Name == columnName
			})

			if colIdx == -1 {
				return s, ColumnNotFoundError(s.EntityName)
			}

			table.Columns[colIdx].Comment = word
		} else {
			idx := slices.IndexFunc(*s.Tables, func(t common.Table) bool {
				return t.Name == s.EntityName
			})

			if idx == -1 {
				return s, TableNotFoundError(s.EntityName)
			}

			(*s.Tables)[idx].Comment = word
		}
		return &NoneState{s.Tables}, nil
	}

	if s.Step == COMMENT_STATE_STEP_IS && isSqlEquals(word, IsWord) {
		s.Step = COMMENT_STATE_STEP_COMMENT_VALUE
	}

	if s.Step == COMMENT_STATE_STEP_ENTITY_NAME {
		s.Step = COMMENT_STATE_STEP_IS
		s.EntityName = word
	}

	if s.Step == COMMENT_STATE_STEP_COLUMN_OR_TABLE && isSqlEquals(word, TableWord) {
		s.Step = COMMENT_STATE_STEP_ENTITY_NAME
		s.IsColumn = false
	}

	if s.Step == COMMENT_STATE_STEP_COLUMN_OR_TABLE && isSqlEquals(word, ColumnWord) {
		s.Step = COMMENT_STATE_STEP_ENTITY_NAME
		s.IsColumn = true
	}

	if s.Step == COMMENT_STATE_STEP_ON && isSqlEquals(word, OnWord) {
		s.Step = COMMENT_STATE_STEP_COLUMN_OR_TABLE

	}

	return s, nil
}
