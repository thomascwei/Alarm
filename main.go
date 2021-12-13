package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"

	"alarm/internal"
)

var (
	Trace *log.Logger
	Info  *log.Logger
	Error *log.Logger
)

// 初始化配置
func init() {
	// log配置
	newPath := filepath.Join(".", "log")
	_ = os.MkdirAll(newPath, os.ModePerm)
	file, err := os.OpenFile("./log/main.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("can not open log file: " + err.Error())
	}
	Trace = log.New(os.Stdout, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(io.MultiWriter(file, os.Stdout), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(file, os.Stdout), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// 讀csv並依指定欄位排序(index從0開始), 後續switch段會需要正確排序
func readCsvFile1(filePath string, SortColNo int) ([][]string, error) {
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

func removeDuplicateValues(StrSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range StrSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)


	AlarmCsv := "./alarm.csv"
	rules, err := readCsvFile1(AlarmCsv, 4)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Trace.Println(rules)
	UniqueIds := make([]string, 0)
	for _, row := range rules {
		UniqueIds = append(UniqueIds, row[0])
		if row[1] == "=" {
			row[1] = "=="
		}
	}
	UniqueIds = removeDuplicateValues(UniqueIds)
	// Trace.Print(UniqueIds)
	ID2Rules := make(map[string][]string)
	for _, row := range rules {
		ID2Rules[row[0]] = []string{}
	}
	for _, row := range rules {
		ID2Rules[row[0]] = append(ID2Rules[row[0]], row[1]+row[2]+":"+row[3])
	}
	// 開始拼湊語法文字
	PackageBase := `	
	package thomas
	import (
		"strconv"
		"fmt"
	)
	`
	for k, v := range ID2Rules {
		//fmt.Println("Key:", k, "values:", v)
		SwitchString := "switch { "
		for _, ele := range v {
			s := strings.Split(ele, ":")
			//log.Print(s)
			SwitchString = SwitchString + "\ncase value " + s[0] + ": " + `return "` + s[1] + `" `
		}
		SwitchString = SwitchString + "\n" + `default:return "pass"}`
		//fmt.Println(SwitchString)

		FunctionBase := `
		func FunctionName(strx string) string{
			fmt.Println("FunctionName got", strx)			
			value, _ := strconv.Atoi(strx)
		`
		FunctionBase = strings.Replace(FunctionBase, "FunctionName", k, 2)
		PackageBase = PackageBase + FunctionBase + SwitchString + "}"
	}
	// 印出組合好的程序
	// Trace.Println(PackageBase)
	// 初始化eval功能
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)
	// 套件程式碼, 照抄
	_, err = i.Eval(PackageBase)
	if err != nil {
		log.Fatal(err)
	}
	funcMap := map[string]func(string) string{}
	for _, objid := range UniqueIds {
		v, err := i.Eval("thomas." + objid)
		if err != nil {
			log.Fatal(err)
		}
		funcMap[objid] = v.Interface().(func(string) string)
	}

	//以Gin框架起一個post接收數據
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "got data",
		})
	})
	r.POST("/V1/InsertSensorData", func(c *gin.Context) {

		ObjectID := c.PostForm("objectID")
		Value := c.PostForm("value")
		currentFunc, ok := funcMap[ObjectID]
		if !ok {
			c.JSON(http.StatusOK, gin.H{
				"message": ObjectID + " not found",
				"Result":  "error",
			})
			return
		}
		result := currentFunc(Value)
		//输出json结果给调用方
		c.JSON(http.StatusOK, gin.H{
			"message":  "got data",
			"ObjectID": ObjectID,
			"Value":    Value,
			"Result":   result,
		})

	})
	go internal.GrpcServer()
	r.Run(":8080")

}
