package internal

import (
	"context"
	"errors"
	"strconv"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"alarm/pkg/db"
	"alarm/pkg/viper"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
)

var (
	file, _ = os.OpenFile("./log/utils.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	Trace   = log.New(os.Stdout, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info    = log.New(io.MultiWriter(file, os.Stdout), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error   = log.New(io.MultiWriter(file, os.Stdout), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	DBconfig     = viper.LoadConfig("./config")
	DBConnection = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true",
		DBconfig.User, DBconfig.Password, DBconfig.Host, DBconfig.Port, DBconfig.DB)
	//DBConnection = "thomas:123456@tcp(host.docker.internal:3306)/schedule?charset=utf8&parseTime=true"
	MyDB, _ = sql.Open("mysql", DBConnection)
	ctx     = context.Background()
	queries = db.New(MyDB)
)

// TODO, 服務初始化調用
func init() {
	// InitSQLAlarmRulesFromCSV("./alarm.csv")
}

// 將alarm rules從csv寫進SQL, 會跳過規格不符的部分將正確的寫進SQL, 失敗的部分會回傳error log
func InitSQLAlarmRulesFromCSV(path string) (err error) {
	// read csv
	f, err := os.Open(path)
	if err != nil {
		Error.Println(err)
		return
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		Error.Println(err)
		return
	}
	records = records[1:]
	sort.Slice(records, func(i, j int) bool {
		return records[i][4] < records[j][4]
	})
	logString := ""
	// write sql
	for _, row := range records {
		Alarmcategoryorder, err := strconv.Atoi(row[4])
		if err != nil {
			Error.Println(err)
			logString += err.Error() + ";"
			continue
		}
		_, err = queries.CreateRule(ctx, db.CreateRuleParams{
			Object:             row[0],
			Alarmcategoryorder: int32(Alarmcategoryorder),
			Alarmlogic:         row[1],
			Triggervalue:       row[2],
			Alarmcategory:      row[3],
			Alamrmessage:       row[5],
		})
		if err != nil {
			Error.Println(err)
			logString += err.Error() + ";"
			continue
		}
	}
	if logString != "" {
		err = errors.New(logString)
	}
	return

}

// TODO,從SQL撈alarm rules, 產生的判斷函數後存進go-cache
func SaveAllFuctionAsCache(filePath string, SortColNo int) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	records = records[1:]
	sort.Slice(records, func(i, j int) bool {
		return records[i][SortColNo] < records[j][SortColNo]
	})

	return records, err
}

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
	if trigger && exist && alarmCategoryCurrent != alarmCategoryCache {
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
func ReceiveAckMessage(objectID string, message string) {
	fmt.Println("")
	// 檢查alarm狀態
	// alarm狀態仍未告警中
	// 寫進cache ack message
	// alarm狀態為已恢復正常則
	// ack message 寫進SQL並補完eventid
	// 刪除cache
}
