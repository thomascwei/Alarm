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

	DBconfig     = viper.LoadConfig(path.Join(rootPath, "config"))
	DBConnection = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&loc=Local",
		DBconfig.User, DBconfig.Password, DBconfig.Host, DBconfig.Port, DBconfig.DB)
	//DBConnection = "thomas:123456@tcp(host.docker.internal:3306)/schedule?charset=utf8&parseTime=true"
	MyDB, _ = sql.Open("mysql", DBConnection)
	ctx     = context.Background()
	queries = db.New(MyDB)
)

const FunctionCacheKey = "funcMap"

// 服務初始化調用
func init() {
	err := SaveAllFuctionAsCache()
	if err != nil {
		log.Fatal(err)
	}
	// 初始化時將SQL內現存未取消alarm存進cache
	LoadActiveAlarmsToCache()
}

// 將SQL內現存未取消alarm存進cache
func LoadActiveAlarmsToCache() {
	events, err := queries.ListAllActiveAlarms(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, event := range events {
		err = GC.Set(event.Object, AlarmCacher{
			Object:                    event.Object,
			EventID:                   int(event.ID),
			AlarmCategoryCurrent:      "",
			AlarmCategoryOrderCurrent: 0,
			AlarmCategoryHigh:         event.Highestalarmcategory,
			AlarmCategoryHighOrder:    int(event.Alarmcategoryorder),
			AlarmMessage:              event.Alarmmessage,
			AckMessage:                event.Ackmessage,
			StartTime:                 event.StartTime,
		})
		if err != nil {
			log.Fatal(err)
		}
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
		AlarmCategoryOrder, err := strconv.Atoi(row[4])
		if err != nil {
			Error.Println(err)
			logString += err.Error() + ";"
			continue
		}
		_, err = queries.CreateRule(ctx, db.CreateRuleParams{
			Object:             row[0],
			Alarmcategoryorder: int32(AlarmCategoryOrder),
			Alarmlogic:         row[1],
			Triggervalue:       row[2],
			Alarmcategory:      row[3],
			Alarmmessage:       row[5],
			Ackmethod:          row[6],
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
	err = SaveAllFuctionAsCache()
	if err != nil {
		log.Fatal(err)
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
		defaultConfig := false
		SwitchString := "switch { "
		for _, element := range v {
			s := strings.Split(element, ":")
			if s[0] == "==others" {
				defaultConfig = true
				continue
			}
			SwitchString = SwitchString + "\ncase value " + s[0] + ": " + `	return []string{"` +
				s[1] + `","` + s[2] + `","` + s[3] + `","` + s[4] + `"}`
		}
		if defaultConfig {
			SwitchString = SwitchString + "\n" + `default:return []string{"pass", "", "", ""}}`
		} else {
			SwitchString = SwitchString + `}; return []string{"undefined", "", "", ""}`
		}

		FunctionBase := `
func FunctionName(strx string) []string{
	value, _ := strconv.ParseFloat(strx,32)
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
			row.Alarmlogic+row.Triggervalue+":"+row.Alarmcategory+":"+row.Alarmmessage+
				":"+strconv.Itoa(int(row.Alarmcategoryorder))+":"+row.Ackmethod)
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
	err = GC.Set(FunctionCacheKey, funcMap)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// 依點位名稱及數值判斷是否觸發alarm
func AlarmTriggerCheck(objectID string, value string) (trigger bool, alarmCategory, alarmMessage string,
	alarmCategoryOrder int, ackMethod string) {
	// 從cache取得對應之function
	funcMapCache, err := GC.Get("funcMap")

	if err != nil {
		Error.Println(err)
		return false, "", "", -1, ""
	}
	funcMap, ok := funcMapCache.(map[string]func(string) []string)
	if !ok {
		Error.Printf("function assert error %T\n", funcMap)
		return false, "", "", -1, ""
	}

	currentFunc, ok := funcMap[objectID]
	// 代表id錯誤, alarmCategory返回-2
	if !ok {
		return false, "", "", -2, ""
	}
	result := currentFunc(value)
	// 未達觸發條件
	if result[0] == "pass" {
		return false, "pass", "", -1, ""
	}
	// 代表sensor送上未定義值
	if result[0] == "undefined" {
		return false, "undefined", "", -3, ""
	}
	order, err := strconv.Atoi(result[2])
	if err != nil {
		Error.Println(err)
		return false, "", "", -1, ""
	}
	return true, result[0], result[1], order, result[3]
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

// 接AlarmTriggerCheck,依觸發結果進行後續邏輯運作
func HandleAlarmTriggerResult(objectID string, value string) (err error) {
	trigger, triggeredAlarmCategory, alarmMessage, alarmCategoryOrder, ackMethod := AlarmTriggerCheck(objectID, value)
	Trace.Println(trigger, triggeredAlarmCategory, alarmCategoryOrder)
	// 若Object id 有誤alarmCategoryOrder為-2
	if alarmCategoryOrder == -2 {
		return errors.New("ObjectID Not Found")
	}
	// 若數值未定義alarmCategoryOrder為-2
	if alarmCategoryOrder == -3 {
		return errors.New("value Not defined")
	}
	exist, AlarmCache := ReadAlarmStatusFromCache(objectID)
	currentTime := time.Now()

	if trigger { // 此次HotData到達觸發標準
		//Trace.Println()
		if exist { //	目前有alarm
			Trace.Println()
			if alarmCategoryOrder == AlarmCache.AlarmCategoryOrderCurrent { // alarm等級不變
				// 新增SQL event detail
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
			} else { //	alarm等級改變
				if alarmCategoryOrder < AlarmCache.AlarmCategoryOrderCurrent { // alarm升級
					// 更新cache,更新SQL event等級與新增SQL event detail
					//	更新cache
					AlarmCache.AlarmCategoryOrderCurrent = alarmCategoryOrder
					AlarmCache.AlarmCategoryCurrent = triggeredAlarmCategory
					AlarmCache.AlarmCategoryHighOrder = alarmCategoryOrder
					AlarmCache.AlarmCategoryHigh = triggeredAlarmCategory
					err = GC.Set(objectID, AlarmCache)
					if err != nil {
						Error.Println(err)
						return err
					}
					//	更新SQL event等級
					err = queries.UpgradeAlarmCategory(ctx, db.UpgradeAlarmCategoryParams{
						Alarmcategoryorder:   int32(alarmCategoryOrder),
						Highestalarmcategory: triggeredAlarmCategory,
						Alarmmessage:         alarmMessage,
						ID:                   int32(AlarmCache.EventID),
					})
					if err != nil {
						Error.Println(err)
						return err
					}
					// 新增SQL event detail
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
				} else { //	alarm降級
					// 	更新cache與新增SQL event detail
					//	更新cache
					AlarmCache.AlarmCategoryOrderCurrent = alarmCategoryOrder
					AlarmCache.AlarmCategoryCurrent = triggeredAlarmCategory
					err = GC.Set(objectID, AlarmCache)
					if err != nil {
						Error.Println(err)
						return err
					}
					// 新增SQL event detail
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
			}
		} else { // 目前無alarm
			// 新增alarm,新增SQL event,cache, 與event detail
			// 產生新alarm, 先寫進SQL並取得eventid後寫進cache
			sqlResult, err := queries.CreateAlarmEvent(ctx, db.CreateAlarmEventParams{
				Object:               objectID,
				Alarmcategoryorder:   int32(alarmCategoryOrder),
				Highestalarmcategory: triggeredAlarmCategory,
				Alarmmessage:         alarmMessage,
				Ackmessage:           ackMethod,
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
				EventID:                   int(eventID),
				AlarmCategoryCurrent:      triggeredAlarmCategory,
				AlarmCategoryOrderCurrent: alarmCategoryOrder,
				AlarmCategoryHigh:         triggeredAlarmCategory,
				AlarmCategoryHighOrder:    alarmCategoryOrder,
				AlarmMessage:              alarmMessage,
				AckMessage:                ackMethod,
				StartTime:                 currentTime})
			if err != nil {
				Error.Println(err)
				return err
			}
		}
	} else { //	此次HotData未達觸發標準
		if exist { // 目前有alarm
			Trace.Println("this")
			if AlarmCache.AckMessage == "" { // 人員未ack
				if alarmCategoryOrder != AlarmCache.AlarmCategoryOrderCurrent { // alarm category不同
					//	更新cache與新增SQL event detail
					//	更新cache
					AlarmCache.AlarmCategoryOrderCurrent = alarmCategoryOrder
					AlarmCache.AlarmCategoryCurrent = triggeredAlarmCategory
					err = GC.Set(objectID, AlarmCache)
					if err != nil {
						Error.Println(err)
						return err
					}
					// 新增SQL event detail
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
				} else { // alarm category相同
					// do nothing
				}
			} else { //人員已ack
				//	刪除cache, 新增SQL event detail, 更新SQL event的endtime
				//	刪除cache
				Trace.Println()
				ok := GC.Remove(objectID)
				if !ok {
					return errors.New("remove cached alarm fail")
				}
				// 新增SQL event detail
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
				//	更新SQL event的endtime
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
		} else { // 目前無alarm
			// do nothing
		}
	}

	return
}

// 接收前端傳來的ack信息寫進cache
func ReceiveAckMessage(objectID string, message string) (err error) {
	// 檢查alarm狀態
	exist, AlarmCache := ReadAlarmStatusFromCache(objectID)
	if !exist {
		return errors.New("this alarm not exist")
	}
	if AlarmCache.AlarmCategoryOrderCurrent == -1 { // alarm狀態為已恢復正常
		// 刪除cache, 更新SQL event的EndTime與AckMessage
		//	刪除cache
		ok := GC.Remove(objectID)
		if !ok {
			return errors.New("remove cached alarm fail")
		}
		// 更新SQL event的AckMessage
		err = queries.UpdateAlarmAckMessage(ctx, db.UpdateAlarmAckMessageParams{
			Ackmessage: message,
			ID:         int32(AlarmCache.EventID),
		})
		if err != nil {
			Error.Println(err)
			return err
		}
		//	更新SQL event的EndTime
		err = queries.SetAlarmEventEndTime(ctx, db.SetAlarmEventEndTimeParams{
			EndTime: sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
			ID: int32(AlarmCache.EventID),
		})
		if err != nil {
			Error.Println(err)
			return err
		}

	} else { // alarm狀態仍在告警中
		// 寫進cache ack message
		AlarmCache.AckMessage = message
		err = GC.Set(objectID, AlarmCache)
		if err != nil {
			Error.Println(err)
			return err
		}
		// 更新SQL AckMessage
		err = queries.UpdateAlarmAckMessage(ctx, db.UpdateAlarmAckMessageParams{
			Ackmessage: message,
			ID:         int32(AlarmCache.EventID),
		})
		if err != nil {
			Error.Println(err)
			return err
		}
	}
	return
}

// 列出cache中的未結alarm清單
func ListAllActiveAlarmsFromCache() (result []AlarmCacher, err error) {
	raw := GC.GetALL(false)
	delete(raw, FunctionCacheKey)
	for _, v := range raw {
		single, ok := v.(AlarmCacher)
		if !ok {
			Error.Println("AlarmCatcher type assert error")
			return nil, errors.New("AlarmCatcher type assert error")
		}
		result = append(result, single)
	}
	return
}
