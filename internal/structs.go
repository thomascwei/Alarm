package internal

import "time"

type AlarmCacher struct {
	Object                    string    `json:"object"`
	EventID                   int64     `json:"event_id"`
	AlarmCategoryCurrent      string    `json:"alar_category"`
	AlarmCategoryOrderCurrent string    `json:"alar_category_order_current"`
	AlarmCategoryHigh         string    `json:"alar_category_high"`
	AlarmMessage              string    `json:"alarm_message"`
	AckMessage                string    `json:"ack_message"`
	StartTime                 time.Time `json:"start_time"`
}
