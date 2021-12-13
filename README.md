# H project

## 模組需求

1. REST server
   - 接收modbus上傳數據
   - 開API對接client應用
2. account
   - 帳密驗證
   - token驗證
3. history
   - sensor data寫進DB
   - 從DB讀sensor data
4. alarm
   - 計算sensor上傳的值是否觸發alarm或解除alarm
   - 產生alarm清單及寫進DB
   - 寫入人員ack的信息

<!-- TODO -->
1. 起gRPC server接Gin server傳來的message,包含sensor data與ack message
2. 判斷是否觸發alarm或解除alarm
3. 將alarm event存cache
4. alarmevent存DB
5. 起gRPC service接收信息, 用於初始化alarm rules(truncate and regen fom csv)

```sh
protoc -I=. --go_out=plugins=grpc:pkg proto/*.proto
```




![](png/ms.png)
