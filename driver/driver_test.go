package driver

import (
	"testing"

	"github.com/icpd/ddl2plantuml/driver/common"
	"github.com/stretchr/testify/assert"
)

func TestTables_Relationship(t *testing.T) {
	t.Run("plural", func(t *testing.T) {
		tables := common.Tables{
			{
				Name:    "users",
				Comment: "user table",
				Columns: []common.Column{{Name: "id"}, {Name: "name"}, {Name: "email"}},
			},
			{
				Name:    "user_detail",
				Comment: "user detail table",
				Columns: []common.Column{{Name: "id"}, {Name: "user_id"}, {Name: "age"}},
			},
		}

		result := tables.Relationship()
		assert.Len(t, result, 1)
		assert.Len(t, result[0].RelationTables, 1)
		assert.Equal(t, "users", result[0].TableName)
		assert.Equal(t, "user_detail", result[0].RelationTables[0].TableName)
		t.Logf("%+v", result)
	})

	t.Run("singular", func(t *testing.T) {
		tables := common.Tables{
			{
				Name:    "user",
				Comment: "user table",
				Columns: []common.Column{{Name: "id"}, {Name: "name"}, {Name: "email"}},
			},
			{
				Name:    "user_detail",
				Comment: "user detail table",
				Columns: []common.Column{{Name: "id"}, {Name: "user_id"}, {Name: "age"}},
			},
		}

		result := tables.Relationship()
		assert.Len(t, result, 1)
		assert.Len(t, result[0].RelationTables, 1)
		assert.Equal(t, "user", result[0].TableName)
		assert.Equal(t, "user_detail", result[0].RelationTables[0].TableName)
		t.Logf("%+v", result)
	})
}
