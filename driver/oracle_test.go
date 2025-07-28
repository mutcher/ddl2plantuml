package driver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOracle_CreateSingleTableParse(t *testing.T) {
	ddl := `CREATE TABLE supplier
(
  supplier_id numeric(10) not null,
  supplier_name varchar2(50) not null,
  contact_name varchar2(50),
  CONSTRAINT supplier_pk PRIMARY KEY (supplier_id)
);
COMMENT ON TABLE supplier IS "hahahah haha haha";`

	d := &Oracle{}
	tables, err := d.Parse(ddl)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tables))
	assert.Equal(t, "supplier", tables[0].Name)
	assert.Equal(t, "hahahah haha haha", tables[0].Comment)
}

func TestOracle_CreateMultipleTableParse(t *testing.T) {
}

func TestOracle_CommentParse(t *testing.T) {
	ddl := `CREATE TABLE something.something (
		id varchar(255)
	);
	COMMENT ON TABLE something.something IS 'Ahaha its a super coment';`

	d := &Oracle{}
	d.Parse(ddl)
	tables, err := d.Parse(ddl)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tables))
	assert.Equal(t, 1, len(tables[0].Columns))
	assert.Equal(t, "something.something", tables[0].Name)
	assert.Equal(t, "Ahaha its a super coment", tables[0].Comment)
}
