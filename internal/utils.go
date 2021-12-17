package internal

import (
	"context"
	"database/sql"
	"errors"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bluele/gcache"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mpvl/unique"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"

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

	_, b, _, _ = runtime.Caller(0)
	rootPath   = filepath.Dir(filepath.Dir(b))

	GC = gcache.New(200).Build()

	// DBconfig     = viper.LoadConfig("./config")
	DBconfig     = viper.LoadConfig(path.Join(rootPath, "config"))
	DBConnection = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true",
		DBconfig.User, DBconfig.Password, DBconfig.Host, DBconfig.Port, DBconfig.DB)
	//DBConnection = "thomas:123456@tcp(host.docker.internal:3306)/schedule?charset=utf8&parseTime=true"
	MyDB, _ = sql.Open("mysql", DBConnection)
	ctx     = context.Background()
	queries = db.New(MyDB)
)

// 服務初始化調用
func init() {
	err := SaveAllFuctionAsCache()
	if err != nil {
		log.Fatal(err)
	}
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

// 產生alarm判定函數的文字檔, 後面會將產出的結果存實體檔觀察
func generateAlarmFunctionString(rules map[string][]string) (PackageBase string) {
	PackageBase = `	
// Code generated by generateAlarmFunctionString. Just for eyeball check
package thomas
import (
		"strconv"
	)
	`
	for k, v := range rules {
		SwitchString := "switch { "
		for _, elememt := range v {
			s := strings.Split(elememt, ":")
			SwitchString = SwitchString + "\ncase value " + s[0] + ": " + `	return []string{"` + s[1] + `","` + s[2] + `","` + s[3] + `"}`
		}
		SwitchString = SwitchString + "\n" + `default:return []string{"pass", "", ""}}`

		FunctionBase := `
func FunctionName(strx string) []string{
	value, _ := strconv.Atoi(strx)
		`
		FunctionBase = strings.Replace(FunctionBase, "FunctionName", k, 2)
		PackageBase = PackageBase + FunctionBase + SwitchString + "}"
	}
	return
}

// 從SQL撈alarm rules, 產生的判斷函數後存進go-cache"funcMap"
func SaveAllFuctionAsCache() (err error) {
	rules, err := queries.ListAllRules(ctx)
	if err != nil {
		return
	}
	UniqueIds := []string{}
	for _, rule := range rules {
		UniqueIds = append(UniqueIds, rule.Object)
	}
	// 必須先排序才能去重複
	sort.Strings(UniqueIds)
	unique.Strings(&UniqueIds)
	// Trace.Println(UniqueIds)

	ID2Rules := make(map[string][]string)
	for _, uid := range UniqueIds {
		ID2Rules[uid] = []string{}
	}
	for _, row := range rules {
		if row.Alarmlogic == "=" {
			row.Alarmlogic = "=="
		}
		ID2Rules[row.Object] = append(ID2Rules[row.Object],
			row.Alarmlogic+row.Triggervalue+":"+row.Alarmcategory+":"+row.Alamrmessage+":"+strconv.Itoa(int(row.Alarmcategoryorder)))
	}
	// Trace.Println(ID2Rules)
	funcString := generateAlarmFunctionString(ID2Rules)
	// 將alarm rules函數存進temp文件夾內, 用於觀察內容是否有誤
	os.WriteFile(path.Join(rootPath, "temp", "temp.go"), []byte(funcString), 0644)

	// 初始化eval功能
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)
	// 套件程式碼, 照抄
	_, err = i.Eval(funcString)
	if err != nil {
		log.Fatal(err)
	}
	funcMap := map[string]func(string) []string{}
	for _, objid := range UniqueIds {
		v, err := i.Eval("thomas." + objid)
		if err != nil {
			log.Fatal(err)
		}
		funcMap[objid] = v.Interface().(func(string) []string)
	}
	// Trace.Println(funcMap)
	err = GC.Set("funcMap", funcMap)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// 依點位名稱及數值判斷是否觸發alarm
func AlarmTriggerCheck(objectID string, value string) (trigger bool, alarmCategory, alarmMessage, alarmCategoryOrder string) {
	// 從cache取得對應之function
	funcMapCache, err := GC.Get("funcMap")

	if err != nil {
		Error.Println(err)
		return false, "", "", ""
	}
	funcMap, ok := funcMapCache.(map[string]func(string) []string)
	if !ok {
		Error.Printf("function assert error %T\n", funcMap)
		return false, "", "", ""
	}

	currentFunc, ok := funcMap[objectID]
	if !ok {
		return false, "", "", ""
	}
	result := currentFunc(value)
	// 未達觸發條件
	if result[0] == "pass" {
		return false, "pass", "", ""
	}
	return true, result[0], result[1], result[2]
}

// 從cache讀此點位當前alarm狀態, 有值返回(true,string), 無值返回(false,"")
func ReadAlarmStatusFromCache(objectID string) (bool, AlarmCacher) {
	resultCache, err := GC.Get(objectID)
	if err != nil {
		Trace.Println(err)
		return false, AlarmCacher{}
	}
	result, ok := resultCache.(AlarmCacher)
	if !ok {
		Error.Println("Parse cache error")
		return false, AlarmCacher{}
	}
	//Trace.Printf("%+v\n", result)
	return true, result
}

// TODO 接AlarmTriggerCheck,依觸發結果進行後續邏輯運作
func HandleAlarmTriggeResult(objectID string, value string) (err error) {
	trigger, triggeredAlarmCategory, alarmMessage, alarmCategoryOrder := AlarmTriggerCheck(objectID, value)
	exist, AlarmCache := ReadAlarmStatusFromCache(objectID)
	currentTime := time.Now()
	// trigger與alarmStatus各有兩種型態形成四個判斷
	// 此次會觸發alarm  目前無alarm
	if trigger && !exist {
		// 產生新alarm, 先寫進SQL並取得eventid後寫進cache
		intalarmCategoryOrder, err := strconv.Atoi(alarmCategoryOrder)
		if err != nil {
			Error.Println(err)
			return err
		}
		sqlResult, err := queries.CreateAlarmEvent(ctx, db.CreateAlarmEventParams{
			Object:               objectID,
			Alarmcategoryorder:   int32(intalarmCategoryOrder),
			Highestalarmcategory: triggeredAlarmCategory,
			Ackmessage:           "",
			StartTime:            currentTime,
		})
		if err != nil {
			Error.Println(err)
			return err
		}
		eventID, err := sqlResult.LastInsertId()
		//Trace.Println(eventID)
		if err != nil {
			Error.Println(err)
			return err
		}
		// 寫進SQL event detail
		_, err = queries.CreateAlarmEventDetail(ctx, db.CreateAlarmEventDetailParams{
			EventID:       int32(eventID),
			Object:        objectID,
			Alarmcategory: triggeredAlarmCategory,
			CreatedAt:     currentTime,
		})
		if err != nil {
			Error.Println(err)
			return err
		}
		// 寫進cache
		err = GC.Set(objectID, AlarmCacher{
			Object:                    objectID,
			EventID:                   eventID,
			AlarmCategoryCurrent:      triggeredAlarmCategory,
			AlarmCategoryOrderCurrent: alarmCategoryOrder,
			AlarmCategoryHigh:         triggeredAlarmCategory,
			AlarmMessage:              alarmMessage,
			AckMessage:                "",
			StartTime:                 currentTime})
		if err != nil {
			Error.Println(err)
			return err
		}
		return nil
	}
	// 到達觸發標準 但目前有alarm且狀態有改變
	if trigger && exist && triggeredAlarmCategory != AlarmCache.AlarmCategoryCurrent {
		//	TODO 如果alarm狀態升級則更新AlarmCategoryHigh

		// 複寫cache中的alarmCategory,SQL event detail新增一筆
		// 複寫cache
		err = GC.Set(objectID, AlarmCacher{
			Object:                    objectID,
			EventID:                   AlarmCache.EventID,
			AlarmCategoryCurrent:      triggeredAlarmCategory,
			AlarmCategoryOrderCurrent: alarmCategoryOrder,
			AlarmCategoryHigh:         AlarmCache.AlarmCategoryHigh,
			AlarmMessage:              alarmMessage,
			AckMessage:                "",
			StartTime:                 currentTime})
		if err != nil {
			Error.Println(err)
			return err
		}
		// 寫進SQL event detail
		_, err = queries.CreateAlarmEventDetail(ctx, db.CreateAlarmEventDetailParams{
			EventID:       int32(AlarmCache.EventID),
			Object:        objectID,
			Alarmcategory: triggeredAlarmCategory,
			CreatedAt:     currentTime,
		})
		if err != nil {
			Error.Println(err)
			return err
		}
	}
	// 此次未達觸發(已正常) 目前仍在告警且未ack
	if !trigger && exist && triggeredAlarmCategory != AlarmCache.AlarmCategoryCurrent {
		// 檢查此alarm是否已ack, 是刪除cache, sql新增一筆並補完eventid, 尚未ack則只複寫cache
		// user尚未ack, 複寫cache & 寫進SQL event detail
		if AlarmCache.AckMessage == "" {
			err = GC.Set(objectID, AlarmCacher{
				Object:                    objectID,
				EventID:                   AlarmCache.EventID,
				AlarmCategoryCurrent:      triggeredAlarmCategory,
				AlarmCategoryOrderCurrent: alarmCategoryOrder,
				AlarmCategoryHigh:         AlarmCache.AlarmCategoryHigh,
				AlarmMessage:              alarmMessage,
				AckMessage:                "",
				StartTime:                 currentTime})
			if err != nil {
				Error.Println(err)
				return err
			}
			// 寫進SQL event detail
			_, err = queries.CreateAlarmEventDetail(ctx, db.CreateAlarmEventDetailParams{
				EventID:       int32(AlarmCache.EventID),
				Object:        objectID,
				Alarmcategory: triggeredAlarmCategory,
				CreatedAt:     currentTime,
			})
		} else {
			Trace.Println("完成")
			// 刪除此筆cache
			GC.Remove(objectID)
			// 寫進SQL event detail
			_, err = queries.CreateAlarmEventDetail(ctx, db.CreateAlarmEventDetailParams{
				EventID:       int32(AlarmCache.EventID),
				Object:        objectID,
				Alarmcategory: triggeredAlarmCategory,
				CreatedAt:     currentTime,
			})
			// SQL Event更新完成時間
			err = queries.SetAlarmEventEndTime(ctx, db.SetAlarmEventEndTimeParams{
				EndTime: sql.NullTime{
					Time:  currentTime,
					Valid: true,
				},
				ID: int32(AlarmCache.EventID),
			})
			if err != nil {
				Error.Println(err)
				return err
			}
		}
	}
	return
}

// TODO 接收前端傳來的ack信息寫進cache
func ReceiveAckMessage(objectID string, message string) {
	fmt.Println("")
	// 檢查alarm狀態
	// alarm狀態仍未告警中
	// 寫進cache ack message
	// alarm狀態為已恢復正常則
	// ack message 寫進SQL並補完eventid
	// 刪除cache
}
