package internal

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInitSQLAlarmRules(t *testing.T) {
	err := InitSQLAlarmRulesFromCSV("./test_alarm.csv")
	require.NoError(t, err)

	var got string
	row := MyDB.QueryRow("SELECT AlarmCategory FROM rules where object=? and AlarmCategoryOrder=?", "ID0012", 2)
	err = row.Scan(&got)
	require.NoError(t, err)
	want := "Medium"
	require.Equal(t, got, want)
}
