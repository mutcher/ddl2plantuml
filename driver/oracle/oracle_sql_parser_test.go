package oracle

import (
	"testing"

	"github.com/icpd/ddl2plantuml/driver/common"
	"github.com/stretchr/testify/assert"
)

type SpySqlHandler struct {
	words []string
}

func (s *SpySqlHandler) GetTables() *common.Tables {
	return &common.Tables{}
}

func (s *SpySqlHandler) InjectWord(word string) error {
	s.words = append(s.words, word)
	return nil
}

func TestWordParser_InjectRightWords1(t *testing.T) {
	// the following string contains spaces and tabs
	ddl := `CREATE TABLE      		  'test 123';`
	d := &Oracle{}
	spy := &SpySqlHandler{}
	d.ParseEx(ddl, spy)

	assert.Equal(t, []string{"CREATE", "TABLE", "test 123", ";"}, spy.words)
}

func TestWordParser_TwoCommandsWithoutComment(t *testing.T) {
	ddl := `CREATE TABLE supplier
(
  supplier_id numeric(10) not null,
  CONSTRAINT supplier_pk PRIMARY KEY (supplier_id)
);
COMMENT ON TABLE supplier IS 'hahahah haha haha';`
	d := &Oracle{}
	spy := &SpySqlHandler{}
	d.ParseEx(ddl, spy)

	assert.Equal(t,
		[]string{
			"CREATE", "TABLE", "supplier", "(",
			"supplier_id", "numeric", "(", "10", ")", "not", "null", ",",
			"CONSTRAINT", "supplier_pk", "PRIMARY", "KEY", "(", "supplier_id", ")", ")", ";",
			"COMMENT", "ON", "TABLE", "supplier", "IS", "hahahah haha haha", ";"},
		spy.words)
}

func TestWordParser_TwoCommands_MultipleKeyConstraint_WithoutComment(t *testing.T) {
	ddl := `CREATE TABLE supplier
(
  supplier_id numeric(10) not null,
  supplier_name varchar(255) not null,
  CONSTRAINT supplier_pk PRIMARY KEY (supplier_id, supplier_name)
);
COMMENT ON TABLE supplier IS 'hahahah haha haha';`
	d := &Oracle{}
	spy := &SpySqlHandler{}
	d.ParseEx(ddl, spy)

	assert.Equal(t,
		[]string{
			"CREATE", "TABLE", "supplier", "(",
			"supplier_id", "numeric", "(", "10", ")", "not", "null", ",",
			"supplier_name", "varchar", "(", "255", ")", "not", "null", ",",
			"CONSTRAINT", "supplier_pk", "PRIMARY", "KEY", "(", "supplier_id", ",", "supplier_name", ")", ")", ";",
			"COMMENT", "ON", "TABLE", "supplier", "IS", "hahahah haha haha", ";"},
		spy.words)
}

func TestWordParser_TwoCommandsWithComment(t *testing.T) {
	ddl := `CREATE TABLE supplier
(
  supplier_id numeric(10) not null,
  /*some comment inside*/
  CONSTRAINT supplier_pk PRIMARY KEY (supplier_id)
);
COMMENT ON TABLE supplier IS 'hahahah haha haha';`
	d := &Oracle{}
	spy := &SpySqlHandler{}
	d.ParseEx(ddl, spy)

	assert.Equal(t,
		[]string{
			"CREATE", "TABLE", "supplier", "(",
			"supplier_id", "numeric", "(", "10", ")", "not", "null", ",",
			"CONSTRAINT", "supplier_pk", "PRIMARY", "KEY", "(", "supplier_id", ")", ")", ";",
			"COMMENT", "ON", "TABLE", "supplier", "IS", "hahahah haha haha", ";"},
		spy.words)
}

func TestWordParser_TwoCommandsWithMultilineComment(t *testing.T) {
	ddl := `CREATE TABLE supplier
(
  supplier_id numeric(10) not null,
  /*some comment inside
  this is multiline commment
  loooong multiline comment   */
  CONSTRAINT supplier_pk PRIMARY KEY (supplier_id)
);
COMMENT ON TABLE supplier IS 'hahahah haha haha';`
	d := &Oracle{}
	spy := &SpySqlHandler{}
	d.ParseEx(ddl, spy)

	assert.Equal(t,
		[]string{
			"CREATE", "TABLE", "supplier", "(",
			"supplier_id", "numeric", "(", "10", ")", "not", "null", ",",
			"CONSTRAINT", "supplier_pk", "PRIMARY", "KEY", "(", "supplier_id", ")", ")", ";",
			"COMMENT", "ON", "TABLE", "supplier", "IS", "hahahah haha haha", ";"},
		spy.words)
}
