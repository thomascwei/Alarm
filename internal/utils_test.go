package internal

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"path"
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
	err := InitSQLAlarmRulesFromCSV(path.Join(rootPath, "test_data", "test_alarm.csv"))
	// 這裡不檢查錯誤
	// require.NoError(t, err)

	var got string
	row := MyDB.QueryRow("SELECT AlarmCategory FROM rules where object=? and AlarmCategoryOrder=?", "ID0012", 2)
	err = row.Scan(&got)
	Trace.Println(got)
	require.NoError(t, err)
	want := "Medium"
	require.Equal(t, got, want)
}

func TestSaveAllFunctionAsCache(t *testing.T) {
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

	_, alarmCategory, _, _, _ := AlarmTriggerCheck(Object, TriggerValue)
	got := alarmCategory
	want := AlarmCategory
	// Trace.Println(got, want)
	require.Equal(t, want, got)

}

func TestReadAlarmStatusFromCache(t *testing.T) {
	object := "ID0011"
	_, alarmCategory, _, _, _ := AlarmTriggerCheck(object, "1")
	Trace.Println(alarmCategory)
	err := GC.Set(object, AlarmCacher{
		Object:                    object,
		EventID:                   11,
		AlarmCategoryCurrent:      alarmCategory,
		AlarmCategoryOrderCurrent: 1,
		AlarmCategoryHigh:         alarmCategory,
		AlarmMessage:              "alarmMessage",
		AckMessage:                "",
		StartTime:                 time.Now()})
	require.NoError(t, err)

	want := "High"

	t.Run("exact Object ID", func(t *testing.T) {
		_, AlarmCache := ReadAlarmStatusFromCache(object)
		require.Equal(t, want, AlarmCache.AlarmCategoryHigh)
	})
	t.Run("wrong Object id", func(t *testing.T) {
		_, AlarmCache := ReadAlarmStatusFromCache("WrongID")
		require.Equal(t, "", AlarmCache.AlarmCategoryHigh)
	})

}

func TestHandleAlarmTriggerResult(t *testing.T) {
	t.Run("trigger High from pass", func(t *testing.T) {
		// 先清空DB
		clearAlarmDBHistory(t)
		object := "ID0012"
		GC.Remove(object)

		err := HandleAlarmTriggerResult(object, "60")
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
		GC.Remove(object)

		// 觸發High
		err := HandleAlarmTriggerResult(object, "80")
		require.NoError(t, err)
		// 觸發Medium
		err = HandleAlarmTriggerResult(object, "50")
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
		GC.Remove(object)

		// 觸發Low
		err := HandleAlarmTriggerResult(object, "21")
		require.NoError(t, err)
		// 回歸正常
		err = HandleAlarmTriggerResult(object, "1")
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
		GC.Remove(object)
		// 觸發Low
		err := HandleAlarmTriggerResult(object, "60")
		require.NoError(t, err)
		err = ReceiveAckMessage(object, "test insert")
		require.NoError(t, err)
		// 回歸正常
		err = HandleAlarmTriggerResult(object, "1")
		require.NoError(t, err)

	})

	t.Run("升升降降先pass再ack", func(t *testing.T) {
		// 先清空DB
		clearAlarmDBHistory(t)
		object := "ID0013"
		GC.Remove(object)

		// 第一階段觸發Low
		want := "Low"
		err := HandleAlarmTriggerResult(object, "11")
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

		//	第二階段觸發Medium
		want = "Medium"
		err = HandleAlarmTriggerResult(object, "50")
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

		//	第三階段觸發High
		want = "High"
		err = HandleAlarmTriggerResult(object, "80")
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

		//	第四階段觸發Low
		err = HandleAlarmTriggerResult(object, "11")
		require.NoError(t, err)
		//	檢測DB歷史最高是否正確
		row = MyDB.QueryRow("SELECT HighestAlarmCategory FROM history_event limit 1")
		err = row.Scan(&got)
		require.NoError(t, err)
		require.Equal(t, "High", got)
		AlarmCache, err = GC.Get(object)
		require.NoError(t, err)
		AlarmCacheAsserted, ok = AlarmCache.(AlarmCacher)
		require.True(t, ok)
		got = AlarmCacheAsserted.AlarmCategoryHigh
		require.Equal(t, "High", got)
		//	檢測當前cache alarmCategory是否正確
		got = AlarmCacheAsserted.AlarmCategoryCurrent
		require.Equal(t, "Low", got)

		//	第五階段觸發pass
		err = HandleAlarmTriggerResult(object, "0")
		require.NoError(t, err)
		//	檢測DB歷史最高是否正確
		row = MyDB.QueryRow("SELECT HighestAlarmCategory FROM history_event limit 1")
		err = row.Scan(&got)
		require.NoError(t, err)
		require.Equal(t, "High", got)
		AlarmCache, err = GC.Get(object)
		require.NoError(t, err)
		AlarmCacheAsserted, ok = AlarmCache.(AlarmCacher)
		require.True(t, ok)
		got = AlarmCacheAsserted.AlarmCategoryHigh
		require.Equal(t, "High", got)
		//	檢測當前cache alarmCategory是否正確
		got = AlarmCacheAsserted.AlarmCategoryCurrent
		require.Equal(t, "pass", got)

		//	第六階段寫進ack後close
		err = ReceiveAckMessage(object, "thomas pass test")
		require.NoError(t, err)
		//	檢測DB歷史最高是否正確
		row = MyDB.QueryRow("SELECT HighestAlarmCategory FROM history_event limit 1")
		err = row.Scan(&got)
		require.NoError(t, err)
		require.Equal(t, "High", got)
		// 檢測ack messages
		row = MyDB.QueryRow("SELECT AckMessage FROM history_event limit 1")
		err = row.Scan(&got)
		require.NoError(t, err)
		require.Equal(t, "thomas pass test", got)
		// 檢測EndTime
		row = MyDB.QueryRow("SELECT count(*) FROM history_event where end_time is null")
		var count int
		err = row.Scan(&count)
		require.Zero(t, count)
		// 檢測cache是否已消失
		AlarmCache, err = GC.Get(object)
		require.Error(t, err)

	})
	t.Run("先ack再pass", func(t *testing.T) {
		//	先觸發Low ,ack, pass
		// 先清空DB
		clearAlarmDBHistory(t)
		object := "ID0013"
		GC.Remove(object)
		// 第一階段觸發Low
		want := "Low"
		err := HandleAlarmTriggerResult(object, "11")
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

		//	第二階段寫進ack後
		err = ReceiveAckMessage(object, "thomas pass test")
		require.NoError(t, err)
		//	檢測DB歷史最高是否正確
		row = MyDB.QueryRow("SELECT HighestAlarmCategory FROM history_event limit 1")
		err = row.Scan(&got)
		require.NoError(t, err)
		require.Equal(t, "Low", got)
		// 檢測ack messages
		row = MyDB.QueryRow("SELECT AckMessage FROM history_event limit 1")
		err = row.Scan(&got)
		require.NoError(t, err)
		require.Equal(t, "thomas pass test", got)
		// 檢測EndTime
		row = MyDB.QueryRow("SELECT count(*) FROM history_event where end_time is null")
		var count int
		err = row.Scan(&count)
		require.Equal(t, 1, count)
		// 檢測cache
		AlarmCache, err = GC.Get(object)
		require.NoError(t, err)
		AlarmCacheAsserted, ok = AlarmCache.(AlarmCacher)
		require.True(t, ok)
		got = AlarmCacheAsserted.AlarmCategoryHigh
		require.Equal(t, "Low", got)
		//	檢測當前cache alarmCategory是否正確
		got = AlarmCacheAsserted.AlarmCategoryCurrent
		require.Equal(t, "Low", got)

		//	第三階段收到pass
		err = HandleAlarmTriggerResult(object, "0")
		require.NoError(t, err)
		//	檢測DB歷史最高是否正確
		row = MyDB.QueryRow("SELECT HighestAlarmCategory FROM history_event limit 1")
		err = row.Scan(&got)
		require.NoError(t, err)
		require.Equal(t, "Low", got)
		// 檢測ack messages
		row = MyDB.QueryRow("SELECT AckMessage FROM history_event limit 1")
		err = row.Scan(&got)
		require.NoError(t, err)
		require.Equal(t, "thomas pass test", got)
		// 檢測EndTime
		row = MyDB.QueryRow("SELECT count(*) FROM history_event where end_time is null")
		err = row.Scan(&count)
		require.Zero(t, count)
		// 檢測cache是否已消失
		AlarmCache, err = GC.Get(object)
		require.Error(t, err)
	})
	t.Run("低到高到pass後自動核銷", func(t *testing.T) {
		//	先觸發Low ,ack, pass
		// 先清空DB
		clearAlarmDBHistory(t)
		object := "ID0015"
		GC.Remove(object)
		// 第一階段觸發Low
		want := "Low"
		err := HandleAlarmTriggerResult(object, "11")
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

		//	第二階段觸發High
		want = "High"
		err = HandleAlarmTriggerResult(object, "80.05")
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

		//	第三階段降到pass後直接結束
		err = HandleAlarmTriggerResult(object, "0.1")
		require.NoError(t, err)
		//	檢測DB歷史最高是否正確
		row = MyDB.QueryRow("SELECT HighestAlarmCategory FROM history_event limit 1")
		err = row.Scan(&got)
		require.NoError(t, err)
		require.Equal(t, "High", got)
		// 檢測EndTime
		var count int
		row = MyDB.QueryRow("SELECT count(*) FROM history_event where end_time is null")
		err = row.Scan(&count)
		require.Zero(t, count)
		// 檢測cache是否已消失
		AlarmCache, err = GC.Get(object)
		require.Error(t, err)
	})
}

// TODO
func TestListAllActiveAlarmsFromCache(t *testing.T) {
	// 先清空DB
	clearAlarmDBHistory(t)
	GC.Remove("ID0012")
	GC.Remove("ID0015")

	err := HandleAlarmTriggerResult("ID0012", "80.05")
	require.NoError(t, err)
	err = HandleAlarmTriggerResult("ID0015", "80.05")
	require.NoError(t, err)
	result, err := ListAllActiveAlarmsFromCache()
	require.NoError(t, err)
	require.Equal(t, 2, len(result))
}
