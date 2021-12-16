package internal

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInitSQLAlarmRules(t *testing.T) {
	InitSQLAlarmRulesFromCSV("./test_alarm.csv")

	var got string
	row := MyDB.QueryRow("SELECT AlarmCategory FROM rules where object=? and AlarmCategoryOrder=?", "ID0012", 2)
	err := row.Scan(&got)
	Trace.Println(got)
	require.NoError(t, err)
	want := "Medium"
	require.Equal(t, got, want)
}

func TestSaveAllFuctionAsCache(t *testing.T) {
	err := SaveAllFuctionAsCache()
	require.NoError(t, err)
	var want int
	row := MyDB.QueryRow("SELECT count(distinct Object) FROM rules ")
	err = row.Scan(&want)
	require.NoError(t, err)
	funcCache, err := GC.Get("funcMap")
	require.NoError(t, err)
	funcMap := funcCache.(map[string]func(string) []string)
	got := len(funcMap)
	require.Equal(t, want, got)
}

func TestAlarmTriggerCheck(t *testing.T) {
	var (
		Object        string
		TriggerValue  string
		AlarmCategory string
		MaxID         int
	)
	// unit test GC需要單獨存
	err := SaveAllFuctionAsCache()
	require.NoError(t, err)
	row := MyDB.QueryRow("SELECT max(id) FROM rules where AlarmLogic='='")
	err = row.Scan(&MaxID)
	require.NoError(t, err)
	// Trace.Println(MaxID)
	err = MyDB.QueryRow("SELECT Object, TriggerValue, AlarmCategory FROM rules WHERE id=?", MaxID).Scan(&Object, &TriggerValue, &AlarmCategory)
	require.NoError(t, err)

	_, alarmCategory, _, _ := AlarmTriggerCheck(Object, TriggerValue)
	got := alarmCategory
	want := AlarmCategory
	// Trace.Println(got, want)
	require.Equal(t, want, got)

}

func TestReadAlarmStatusFromCache(t *testing.T) {
	object := "aa"
	want := "High"
	GC.Set(object, []string{want, "ttyytest", ""})
	t.Run("exact objectid", func(t *testing.T) {
		_, got, _, _ := ReadAlarmStatusFromCache(object)
		require.Equal(t, want, got)
	})
	t.Run("wrong objectid", func(t *testing.T) {
		_, got, _, _ := ReadAlarmStatusFromCache("wrongid")
		require.Equal(t, "", got)
	})

}
