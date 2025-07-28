package driver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSqlStateHandlerImpl_CreateSimpleTable_NoComments(t *testing.T) {
	words := []string{"CREATE", "TABLE", "suppa_table", "(", "id", "varchar", "(", "255", ")", ")", ";"}
	wordsProcessor := SqlStateHandlerImpl{}
	wordsProcessor.Reset()

	var err error

	for _, word := range words {
		err := wordsProcessor.InjectWord(word)
		if err != nil {
			break
		}
	}

	tables := wordsProcessor.GetTables()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(*tables))
	tmpTable := (*tables)[0]
	assert.Equal(t, "suppa_table", tmpTable.Name)
	assert.Equal(t, 1, len(tmpTable.Columns))
	assert.Equal(t, "id", tmpTable.Columns[0].Name)
	assert.Equal(t, "varchar(255)", tmpTable.Columns[0].Type)
}

func TestSqlStateHandlerImpl_CreateSimpleTable_SkipCommandsExceptCreateaAncComment(t *testing.T) {
	words := []string{"CREATE", "TABLE", "suppa_table", "(", "id", "varchar", "(", "255", ")", ")", ";"}
	wordsProcessor := SqlStateHandlerImpl{}
	wordsProcessor.Reset()

	var err error

	for _, word := range words {
		err := wordsProcessor.InjectWord(word)
		if err != nil {
			break
		}
	}

	tables := wordsProcessor.GetTables()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(*tables))
	tmpTable := (*tables)[0]
	assert.Equal(t, "suppa_table", tmpTable.Name)
	assert.Equal(t, 1, len(tmpTable.Columns))
	assert.Equal(t, "id", tmpTable.Columns[0].Name)
	assert.Equal(t, "varchar(255)", tmpTable.Columns[0].Type)
}

func TestSqlStateHandlerImpl_CreateSimpleTable_ThrowsErrorWithIncorrectSyntax(t *testing.T) {
	words := []string{"CREATE", "IS", "TABLE", "suppa_table", "(", "id", "varchar(255)", ")", ";"}
	wordsProcessor := SqlStateHandlerImpl{}
	wordsProcessor.Reset()

	var errorGiven error

	for _, word := range words {
		err := wordsProcessor.InjectWord(word)
		if err != nil {
			errorGiven = err
			break
		}
	}

	assert.Equal(t, "Unexpected word in sequence IS prev word: CREATE", errorGiven.Error())
}

func TestSqlStateHandlerImpl_CreateSimpleTable_WithComments(t *testing.T) {
	words := []string{
		"CREATE", "TABLE", "suppa_table", "(", "id", "varchar", "(", "255", ")", ")", ";",
		"COMMENT", "ON", "TABLE", "suppa_table", "IS", "this is potuzhyi comment", ";",
	}
	wordsProcessor := SqlStateHandlerImpl{}
	wordsProcessor.Reset()

	var err error

	for _, word := range words {
		err := wordsProcessor.InjectWord(word)
		if err != nil {
			break
		}
	}

	tables := wordsProcessor.GetTables()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(*tables))
	tmpTable := (*tables)[0]
	assert.Equal(t, "suppa_table", tmpTable.Name)
	assert.Equal(t, 1, len(tmpTable.Columns))
	assert.Equal(t, "id", tmpTable.Columns[0].Name)
	assert.Equal(t, "varchar(255)", tmpTable.Columns[0].Type)
}

func TestSqlStateHandlerImpl_CreateMultipleTables_NoComments(t *testing.T) {
}

func TestSqlStateHandlerImpl_CreateMultipleTables_WithComments(t *testing.T) {
}

func TestCreateTableColumnsDefinitionSubState_ColumnDefinition(t *testing.T) {
	table := Table{}
	state := CreateTableColumnsDefinitionState{&table, Column{}, COLUMN_NAME_STEP}

	words := []string{"id", "varchar(255)", ","}

	for _, word := range words {
		_, err := state.InjectWord(word)
		assert.Nil(t, err)
	}

	assert.Equal(t, "id", state.TmpColumn.Name)
	assert.Equal(t, "varchar(255)", state.TmpColumn.Type)
}
