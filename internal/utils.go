package internal

import "fmt"

// TODO, 服務初始化調用
func init() {
	SaveAllFuctionAsCache()
}

// TODO,將點位csv產生的判斷函數存進go-cache
func SaveAllFuctionAsCache() {}

// TODO,依點位名稱及數值判斷是否觸發alarm
func AlarmTriggerCheck(objectID string, value interface{}) (trigger bool, alarmCategory string) {
	return
}

// TODO,從cache讀此點位當前alarm狀態, 有值返回(true,string), 無值返回(false,"")
func ReadAlarmStatusFromCache(objectID string) (exist bool, alarmCategory string) {
	return
}

// TODO,接AlarmTriggerCheck,依觸發結果進行後續邏輯運作
func HandleAlarmTriggeResult(objectID string, value interface{}) {
	trigger, alarmCategoryCurrent := AlarmTriggerCheck(objectID, value)
	exist, alarmCategoryCache := ReadAlarmStatusFromCache(objectID)
	// trigger與alarmStatus各有兩種型態形成四個判斷
	// 此次會觸發alarm  目前無alarm
	if trigger && !exist {
		// 產生新alarm, 寫進cache與sql
		fmt.Println("")
	}
	// 此次會觸發alarm  目前有alarm且狀態有改變
	if trigger && exist && alarmCategoryCurrent!=alarmCategoryCache{
		// 複寫cache中的alarmCategory,SQL新增一筆
		fmt.Println("")
	}
	// 此次不會觸發 目前有alarm
	if !trigger && exist {
		// 檢查此alarm是否已ack, 是刪除cache, sql新增一筆並補完eventid, 尚未ack則只複寫cache
		fmt.Println("")
	}
}

// TODO,接收前端傳來的ack信息寫進cache
func ReceiveAckMessage(objectID string,message string){
	fmt.Println("")
	// 檢查alarm狀態
		// alarm狀態仍未告警中
			// 寫進cache ack message
		// alarm狀態為已恢復正常則
			// ack message 寫進SQL並補完eventid
			// 刪除cache
}

