package internal

import "time"

type AlarmCacher struct {
	Object                    string    `json:"object"`
	EventID                   int64     `json:"event_id"`
	AlarmCategoryCurrent      string    `json:"alar_category"`
	AlarmCategoryOrderCurrent int       `json:"alar_category_order_current"`
	AlarmCategoryHigh         string    `json:"alar_category_high"`
	AlarmCategoryHighOrder    int       `json:"alarm_category_high_order"`
	AlarmMessage              string    `json:"alarm_message"`
	AckMessage                string    `json:"ack_message"`
	StartTime                 time.Time `json:"start_time"`
}
