package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"google.golang.org/grpc"
	"YourProjectName/proto"
	"log"
	"net"
	"os"
	"sort"
	"strings"
)

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

var channels = make(map[string]chan string)

type server1 struct{}

func (s *server1) Insert(ctx context.Context, in *proto.HotDataRequest) (*proto.HotDataResponse, error) {
	channels[in.ObjectID] <- in.Value
	msg := "Got " + in.ObjectID + ": " + in.Value
	return &proto.HotDataResponse{Message: msg}, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	AlarmCsv := "./alarm.csv"
	rules, err := readCsvFile1(AlarmCsv, 4)
	if err != nil {
		fmt.Println(err)
		return
	}
	//log.Print(rules)
	for _, row := range rules {
		if row[1] == "=" {
			row[1] = "=="
		}
	}
	//log.Print(rules)
	ID2Rules := make(map[string][]string)
	for _, row := range rules {
		ID2Rules[row[0]] = []string{}
	}
	//log.Print(ID2Rules)
	for _, row := range rules {
		ID2Rules[row[0]] = append(ID2Rules[row[0]], row[1]+row[2]+":"+row[3])
	}
	//log.Print(ID2Rules)
	// 開始拼湊語法文字
	PackageBase := `	
	package thomas
	import (
		"fmt"
		"strconv"
	)
	// 送出判斷結果, 這裡只用print, 真實業務場景是往Kafka送
	func SendResult(ObjectID string, value interface{}, res string) {
		fmt.Println(ObjectID, "got:", value, "; send result:", res, "to destination")
	}
	`
	for k, v := range ID2Rules {
		//fmt.Println("Key:", k, "values:", v)
		SwitchString := "switch { "
		for _, ele := range v {
			s := strings.Split(ele, ":")
			//log.Print(s)
			SwitchString = SwitchString + "\ncase value " + s[0] + ": " + `SendResult(ObjectID, value, "` + s[1] + `") `
		}
		SwitchString = SwitchString + "}"
		//fmt.Println(SwitchString)

		FunctionBase := `
		func FunctionName(inchan chan string) {
			ObjectID := "FunctionName"
			for {
				strx := <-inchan
				//fmt.Println("in", strx)
				value, _ := strconv.Atoi(strx)
		`
		FunctionBase = strings.Replace(FunctionBase, "FunctionName", k, 2)
		PackageBase = PackageBase + FunctionBase + SwitchString + "}}"
	}
	// 印出組合好的程序
	//fmt.Println(PackageBase)
	// 初始化eval功能
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)
	// 套件程式碼, 照抄
	_, err = i.Eval(PackageBase)
	if err != nil {
		panic(err)
	}
	// 每個點位都有自己的channel, 以map形式組合
	ObjectIDs := make([]string, 0, len(ID2Rules))
	for k := range ID2Rules {
		ObjectIDs = append(ObjectIDs, k)
	}
	//log.Println(ObjectIDs)
	//channels := make(map[string]chan string)
	// 每個點位的rule生成的function,帶入自己專屬的channel後以goroutine開起來
	for _, funcc := range ObjectIDs {
		channels[funcc] = make(chan string, 10)
		//log.Println(funcc)
		// 用package.function來建一個function
		v, err := i.Eval("thomas." + funcc)
		if err != nil {
			panic(err)
		}
		// 驗證function參數格式
		Calc := v.Interface().(func(chan string))
		// 協程開起來
		go Calc(channels[funcc])
	}

	l, err := net.Listen("tcp", ":55555") // Starts a TCP server listening on port 55555 and handles any errors.
	// The gRPC server will use it.
	if err != nil {
		log.Fatalf("failed to listen for tcp: %s", err)
	}
	s := grpc.NewServer() // Creates a gRPC server and handles requests over the TCP connection
	proto.RegisterHotDataReceiverServer(s, &server1{})

	s.Serve(l) // Registers the implementation of the service on the RPC server.
}