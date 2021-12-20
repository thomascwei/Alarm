package internal

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// 方便測試用, 讓每個unit test都在初始化狀態
func clearAlarmDBHistory(t *testing.T) {
	_, err := MyDB.Exec("SET FOREIGN_KEY_CHECKS = 0;")
	require.NoError(t, err)
	_, err = MyDB.Exec("TRUNCATE history_event;")
	require.NoError(t, err)
	_, err = MyDB.Exec("TRUNCATE history_event_detail")
	require.NoError(t, err)
	_, err = MyDB.Exec("SET FOREIGN_KEY_CHECKS = 1;")
	require.NoError(t, err)
}

func TestInitSQLAlarmRules(t *testing.T) {
	err := InitSQLAlarmRulesFromCSV("./test_alarm.csv")
	//require.NoError(t, err)

	var got string
	row := MyDB.QueryRow("SELECT AlarmCategory FROM rules where object=? and AlarmCategoryOrder=?", "ID0012", 2)
	err = row.Scan(&got)
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
	object := "ID0011"
	_, alarmCategory, _, _ := AlarmTriggerCheck(object, "1")
	Trace.Println(alarmCategory)
	err := GC.Set(object, AlarmCacher{
		Object:                    object,
		EventID:                   11,
		AlarmCategoryCurrent:      alarmCategory,
		AlarmCategoryOrderCurrent: "1",
		AlarmCategoryHigh:         alarmCategory,
		AlarmMessage:              "alarmMessage",
		AckMessage:                "",
		StartTime:                 time.Now()})
	require.NoError(t, err)

	want := "High"

	t.Run("exact objectid", func(t *testing.T) {
		_, AlarmCache := ReadAlarmStatusFromCache(object)
		require.Equal(t, want, AlarmCache.AlarmCategoryHigh)
	})
	t.Run("wrong objectid", func(t *testing.T) {
		_, AlarmCache := ReadAlarmStatusFromCache("wrongid")
		require.Equal(t, "", AlarmCache.AlarmCategoryHigh)
	})

}

func TestHandleAlarmTriggeResult(t *testing.T) {
	t.Run("trigger High from pass", func(t *testing.T) {
		// 先清空DB
		clearAlarmDBHistory(t)

		object := "ID0012"
		err := HandleAlarmTriggeResult(object, "60")
		require.NoError(t, err)
		want := "High"
		var got string
		row := MyDB.QueryRow("SELECT HighestAlarmCategory FROM history_event order by id desc limit 1")
		err = row.Scan(&got)
		require.NoError(t, err)
		require.Equal(t, want, got)
		vv, err := GC.Get(object)
		require.NoError(t, err)
		vvv, ok := vv.(AlarmCacher)
		require.True(t, ok)
		want = vvv.AlarmCategoryCurrent
		require.Equal(t, want, got)

	})
	t.Run("trigger Medium from High", func(t *testing.T) {
		// 先清空DB
		clearAlarmDBHistory(t)

		object := "ID0013"
		// 觸發High
		err := HandleAlarmTriggeResult(object, "80")
		require.NoError(t, err)
		// 觸發Medium
		err = HandleAlarmTriggeResult(object, "50")
		require.NoError(t, err)
		want := "High"
		var got string
		row := MyDB.QueryRow("SELECT HighestAlarmCategory FROM history_event order by id desc limit 1")
		err = row.Scan(&got)
		require.NoError(t, err)
		// DB HighestAlarmCategory will not change
		require.Equal(t, want, got)
		vv, err := GC.Get(object)
		require.NoError(t, err)
		vvv, ok := vv.(AlarmCacher)
		require.True(t, ok)
		got = vvv.AlarmCategoryHigh
		require.Equal(t, want, got)
		// 測cache的current狀態
		want = "Medium"
		got = vvv.AlarmCategoryCurrent
		require.Equal(t, want, got)
	})
	t.Run("Down grade from Low to pass", func(t *testing.T) {
		// 先清空DB
		clearAlarmDBHistory(t)

		object := "ID0014"
		// 觸發Low
		err := HandleAlarmTriggeResult(object, "21")
		require.NoError(t, err)
		// 回歸正常
		err = HandleAlarmTriggeResult(object, "1")
		require.NoError(t, err)

		want := "pass"
		vv, err := GC.Get(object)
		require.NoError(t, err)
		vvv, ok := vv.(AlarmCacher)
		require.True(t, ok)
		got := vvv.AlarmCategoryCurrent
		require.Equal(t, want, got)
	})
	t.Run("解除警報", func(t *testing.T) {
		// 先清空DB
		clearAlarmDBHistory(t)

		object := "ID0012"
		// 觸發Low
		err := HandleAlarmTriggeResult(object, "11")
		// 刻意修改ackmessage
		tempAlarmCache, err := GC.Get(object)
		require.NoError(t, err)
		AlarmCache := tempAlarmCache.(AlarmCacher)
		AlarmCache.AckMessage = "test insert"
		// 寫進cache
		err = GC.Set(object, AlarmCache)
		require.NoError(t, err)

		// 回歸正常
		err = HandleAlarmTriggeResult(object, "1")
		require.NoError(t, err)

	})
	t.Run("升升降降先pass再ack", func(t *testing.T) {
		// 先清空DB
		clearAlarmDBHistory(t)
		object := "ID0013"

		// 第一階段觸發Low
		want := "Low"
		err := HandleAlarmTriggeResult(object, "11")
		require.NoError(t, err)
		//	檢測DB歷史最高是否正確
		var got string
		row := MyDB.QueryRow("SELECT HighestAlarmCategory FROM history_event limit 1")
		err = row.Scan(&got)
		require.NoError(t, err)
		require.Equal(t, want, got)
		AlarmCache, err := GC.Get(object)
		require.NoError(t, err)
		AlarmCacheAsserted, ok := AlarmCache.(AlarmCacher)
		require.True(t, ok)
		got = AlarmCacheAsserted.AlarmCategoryHigh
		require.Equal(t, want, got)
		//	檢測當前cache alarmCategory是否正確
		got = AlarmCacheAsserted.AlarmCategoryCurrent
		require.Equal(t, want, got)

		//	TODO 第二階段觸發Medium
		want = "Medium"
		err = HandleAlarmTriggeResult(object, "50")
		require.NoError(t, err)
		row = MyDB.QueryRow("SELECT HighestAlarmCategory FROM history_event limit 1")
		err = row.Scan(&got)
		require.NoError(t, err)
		require.Equal(t, want, got)
		AlarmCache, err = GC.Get(object)
		require.NoError(t, err)
		AlarmCacheAsserted, ok = AlarmCache.(AlarmCacher)
		require.True(t, ok)
		got = AlarmCacheAsserted.AlarmCategoryHigh
		require.Equal(t, want, got)
		//	檢測當前cache alarmCategory是否正確
		got = AlarmCacheAsserted.AlarmCategoryCurrent
		require.Equal(t, want, got)

		//	TODO 第三階段觸發High
		//	TODO 第四階段觸發Low
		//	TODO 第五階段沒觸發pass
		//	TODO 第六階段寫進ack後close
	})
	t.Run("升升降降先ack再pass", func(t *testing.T) {
		//	TODO 先觸發Low 再觸發High 再觸發Medium, 檢查每次的結果, cache SQL都要
	})
}
