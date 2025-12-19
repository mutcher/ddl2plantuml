package oracle

import (
	"fmt"
	"slices"
	"strings"

	"github.com/icpd/ddl2plantuml/driver/common"
)

const (
	RUNE_FORWARD_SLASH = '/'
	RUNE_STAR          = '*'
)

type SqlStateHandler interface {
	GetTables() *common.Tables
	InjectWord(word string) error
}

type Oracle struct {
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

func (m *Oracle) Parse(ddl string) (common.Tables, error) {
	sqlHandler := new(SqlStateHandlerImpl)
	sqlHandler.Reset()
	tables, err := m.ParseEx(ddl, sqlHandler)
	if err != nil {
		return nil, err
	}

	// deduplicate columns by name per table, preserve order
	for ti := range tables {
		seen := make(map[string]bool)
		filtered := make([]common.Column, 0, len(tables[ti].Columns))
		for _, c := range tables[ti].Columns {
			name := c.Name
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = true
			filtered = append(filtered, c)
		}
		tables[ti].Columns = filtered
	}

	return tables, nil
}

func (m *Oracle) ParseEx(ddl string, sqlHandler SqlStateHandler) (common.Tables, error) {
	var wordBuilder strings.Builder

	wordBuilder.Reset()
	lexer := common.Lexer{}
	lexer.Reset(&ddl)

	for lexer.IsValid() {
		// skip comments
		// skip block comments /* ... */
		if lexer.Current() == RUNE_FORWARD_SLASH && lexer.PreviewNext() == RUNE_STAR {
			lexer.Inc() // skipping "star"
			for {
				lexer.Inc()
				if lexer.Current() == RUNE_STAR && lexer.Next() == RUNE_FORWARD_SLASH {
					break
				}
			}

			lexer.Inc()
			continue
		}

		// skip single-line comments -- until newline
		if lexer.Current() == '-' && lexer.PreviewNext() == '-' {
			// advance until newline or end
			for lexer.IsValid() {
				if lexer.Current() == '\n' {
					break
				}
				lexer.Inc()
			}
			lexer.Inc()
			continue
		}

		// process quotes
		if lexer.Current() == '\'' {
			for {
				nextRune := lexer.Next()
				if nextRune == '\'' {
					break
				}

				wordBuilder.WriteRune(nextRune)
			}
			err := sqlHandler.InjectWord(wordBuilder.String())
			if err != nil {
				return nil, err
			}
			wordBuilder.Reset()

			lexer.Inc()
			continue
		}

		stopRunes := []rune{' ', '\t', '\r', '\n', ',', ';', '(', ')'}
		stopRunesToInclude := []rune{',', ';', '(', ')'}

		if !slices.Contains(stopRunes, lexer.Current()) {
			_, err := wordBuilder.WriteRune(lexer.Current())
			if err != nil {
				return nil, err
			}
		} else {
			if wordBuilder.Len() != 0 {
				err := sqlHandler.InjectWord(wordBuilder.String())
				if err != nil {
					return nil, err
				}
				wordBuilder.Reset()
			}

			if slices.Contains(stopRunesToInclude, lexer.Current()) {
				_, err := wordBuilder.WriteRune(lexer.Current())
				if err != nil {
					return nil, err
				}
				err = sqlHandler.InjectWord(wordBuilder.String())
				if err != nil {
					return nil, err
				}
				wordBuilder.Reset()
			}
		}

		lexer.Inc()
	}

	if wordBuilder.Len() != 0 {
		err := sqlHandler.InjectWord(wordBuilder.String())
		if err != nil {
			return nil, err
		}
	}

	return *sqlHandler.GetTables(), nil
}
