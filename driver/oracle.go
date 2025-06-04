package driver

import (
	"fmt"
	"strings"
)

type Oracle struct {
	Tables Tables
}

// The parser must be based on
// https://docs.oracle.com/en/database/oracle/oracle-database/23/sqlrf/CREATE-TABLE.html
// https://docs.oracle.com/en/database/oracle/oracle-database/23/sqlrf/COMMENT.html

type TableNotFoundError string

func (e TableNotFoundError) Error() string {
	return fmt.Sprintf("Table not found: %s", string(e))
}

type OracleSQLSyntaxError string

func (e OracleSQLSyntaxError) Error() string {
	return fmt.Sprintf("Check the command syntax: %s", string(e))
}

type InvalidTokenError struct {
	Token   string
	Command string
}

func (e InvalidTokenError) Error() string {
	return fmt.Sprintf("Invalid token in statement %s, full: %s", e.Token, e.Command)
}

func isSqlEquals(command string, expected string) bool {
	return strings.ToUpper(strings.TrimSpace(command)) == expected
}

// Searches the table by name, If not found returns -1
func (m *Oracle) FindTableIndex(tableName string) int {
	for i, t := range m.Tables {
		if t.Name == tableName {
			return i
		}
	}
	return -1
}

func (m *Oracle) ParseCommentTableCommand(words []string, commandText string) error {
	tableName := words[0]
	if !isSqlEquals(words[1], "IS") {
		return InvalidTokenError{words[1], commandText}
	}

	comment := strings.Trim(strings.Join(words[2:], " "), "\"")
	i := m.FindTableIndex(tableName)
	if i == -1 {
		return TableNotFoundError(tableName)
	}
	m.Tables[i].Comment = comment

	return nil
}

func (m *Oracle) ParseCommentColumnCommand(words []string, commandText string) error {
	//tableWithColumnName := words[0]
	if !isSqlEquals(words[1], "IS") {
		return InvalidTokenError{words[1], commandText}
	}
	return nil
}

func (m *Oracle) ParseCommentCommand(words []string, commandText string) error {
	// COMMENT keyword already processed
	if !isSqlEquals(words[0], "ON") {
		return InvalidTokenError{words[0], commandText}
	}

	if isSqlEquals(words[1], "TABLE") {
		return m.ParseCommentTableCommand(words[2:], commandText)
	} else if isSqlEquals(words[1], "COLUMN") {
		return m.ParseCommentColumnCommand(words[2:], commandText)
	}

	return nil
}

func (m *Oracle) ParseCreateCommand(words []string, commandText string) error {
	if !isSqlEquals(words[0], "TABLE") {
		// skipping command
		// we're parsing only tables at the moment
		return nil
	}

	table := Table{Name: words[1]}
	m.Tables = append(m.Tables, table)

	return nil
}

func (m *Oracle) SkipCommand(words []string, _ string) error {
	// stub command that does nothing
	// used to make code more accurate
	return nil
}

func (m *Oracle) Parse(ddl string) (Tables, error) {
	m.Tables = Tables{}
	commands := strings.Split(ddl, ";")

	for _, fullCommand := range commands {
		// getting rid of tabs
		fullCommand = strings.ReplaceAll(fullCommand, "\t", " ")
		// getting rid of newlines: linux
		fullCommand = strings.ReplaceAll(fullCommand, "\n", " ")
		// getting rid of newlines: windows style as well
		fullCommand = strings.ReplaceAll(fullCommand, "\r", " ")

		// getting rid of all extra spaces we're generated
		for {
			before := fullCommand
			after := strings.ReplaceAll(before, "  ", " ")
			fullCommand = after

			if before == after {
				break
			}
		}

		fullCommand = strings.Trim(fullCommand, " ")
		words := strings.Split(fullCommand, " ")
		cmdProcessor := m.SkipCommand

		if isSqlEquals(words[0], "CREATE") {
			cmdProcessor = m.ParseCreateCommand
		} else if isSqlEquals(words[0], "COMMENT") {
			cmdProcessor = m.ParseCommentCommand
		}

		err := cmdProcessor(words[1:], fullCommand)
		if err != nil {
			return nil, err
		}
	}
	return m.Tables, nil
}
