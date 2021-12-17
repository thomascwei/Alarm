// Code generated by generateAlarmFunctionString. Just for eyeball check
package thomas

import (
	"strconv"
)

func ID003(strx string) []string {
	value, _ := strconv.Atoi(strx)
	switch {
	case value > 71:
		return []string{"High", "ID003告警1", "1"}
	case value > 41:
		return []string{"Medium", "ID003告警2", "2"}
	case value > 1:
		return []string{"Low", "ID003告警3", "3"}
	default:
		return []string{"pass", "", ""}
	}
}
func ID004(strx string) []string {
	value, _ := strconv.Atoi(strx)
	switch {
	case value == 71:
		return []string{"High", "ID004告警11", "1"}
	case value == 21:
		return []string{"Low", "ID004告警32", "2"}
	case value == 11:
		return []string{"Low", "ID004告警31", "3"}
	default:
		return []string{"pass", "", ""}
	}
}
func ID001(strx string) []string {
	value, _ := strconv.Atoi(strx)
	switch {
	case value == 1:
		return []string{"High", "ID001告警1", "1"}
	case value == 2:
		return []string{"Medium", "ID001告警2", "2"}
	case value == 3:
		return []string{"Low", "ID001告警3", "3"}
	default:
		return []string{"pass", "", ""}
	}
}
func ID0011(strx string) []string {
	value, _ := strconv.Atoi(strx)
	switch {
	case value == 1:
		return []string{"High", "ID001告警1", "1"}
	case value == 2:
		return []string{"Medium", "ID001告警2", "2"}
	case value == 3:
		return []string{"Low", "ID001告警3", "3"}
	default:
		return []string{"pass", "", ""}
	}
}
func ID0012(strx string) []string {
	value, _ := strconv.Atoi(strx)
	switch {
	case value > 50:
		return []string{"High", "ID002告警1", "1"}
	case value > 30:
		return []string{"Medium", "ID002告警2", "2"}
	case value > 10:
		return []string{"Low", "ID002告警3", "3"}
	default:
		return []string{"pass", "", ""}
	}
}
func ID0013(strx string) []string {
	value, _ := strconv.Atoi(strx)
	switch {
	case value > 71:
		return []string{"High", "ID003告警1", "1"}
	case value > 41:
		return []string{"Medium", "ID003告警2", "2"}
	case value > 1:
		return []string{"Low", "ID003告警3", "3"}
	default:
		return []string{"pass", "", ""}
	}
}
func ID0014(strx string) []string {
	value, _ := strconv.Atoi(strx)
	switch {
	case value == 21:
		return []string{"Low", "ID004告警32", "2"}
	case value == 11:
		return []string{"Low", "ID004告警31", "3"}
	default:
		return []string{"pass", "", ""}
	}
}
func ID002(strx string) []string {
	value, _ := strconv.Atoi(strx)
	switch {
	case value > 50:
		return []string{"High", "ID002告警1", "1"}
	case value > 30:
		return []string{"Medium", "ID002告警2", "2"}
	case value > 10:
		return []string{"Low", "ID002告警3", "3"}
	default:
		return []string{"pass", "", ""}
	}
}
