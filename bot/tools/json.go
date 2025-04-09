package tools

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Json struct {
	Data any `json:"data"`
}

func StringToJson(data string) (*Json, error) {
	var jsonData Json

	err := json.Unmarshal([]byte(data), &jsonData.Data)
	if err != nil {
		return &Json{}, fmt.Errorf("failed to unmarshal JSON data!\n%w", err)
	}

	return &jsonData, nil
}

func (js *Json) Get(key string) *Json {
	if value, ok := js.Map()[key]; ok {
		return &Json{value}
	}

	return &Json{}
}

func (js *Json) Exist(key string) bool {
	if _, ok := js.Map()[key]; ok {
		return true
	}

	return false
}

func (js *Json) Index(index int) *Json {
	ja := js.Array()
	length := len(ja)

	if length > index {
		if index >= 0 && index < length {
			return &Json{ja[index]}
		} else if index < 0 && -index <= length {
			return &Json{ja[length+index]}
		}
	}

	return &Json{}
}

func (js *Json) JsonArray() []*Json {
	jsa := []*Json{}

	for _, item := range js.Array() {
		jsa = append(jsa, &Json{item.(map[string]any)})
	}

	return jsa
}

func (js *Json) SubArray(start, end int) []*Json {
	jsa := []*Json{}
	items := js.Array()

	for i, item := range items {
		if i < start {
			continue
		} else if i >= len(items)+end {
			break
		}
		jsa = append(jsa, &Json{item.(map[string]any)})
	}

	return jsa
}

func (js *Json) Map() map[string]any {
	if jm, ok := (js.Data).(map[string]any); ok {
		return jm
	}

	return map[string]any{}
}

func (js *Json) Array() []any {
	if ja, ok := (js.Data).([]any); ok {
		return ja
	}

	return []any{}
}

func (js *Json) Slice(start, end int) string {
	if str, ok := (js.Data).(string); ok {
		if len(str) == 0 {
			return ""
		}

		if start == -1 {
			return str[:end]
		}

		if end == -1 {
			return str[start:]
		}

		return str[start:end]
	}

	return ""
}

func (js *Json) Replace(old, new string, n int) string {
	if str, ok := (js.Data).(string); ok {
		return strings.Replace(str, old, new, n)
	}

	return ""
}

func (js *Json) HasPrefix(sub string) bool {
	return strings.HasPrefix(js.String(), sub)
}

func (js *Json) Split(sep string) []string {
	if str, ok := (js.Data).(string); ok {
		return strings.Split(str, sep)
	}

	return []string{}
}

func (js *Json) Image() string {
	for _, size := range []string{"maxres", "standard", "high", "medium", "default"} {
		if js.Exist(size) {
			s := strings.Replace(js.Get(size).Get("url").String(), "_live.", ".", 1)

			if strings.Contains(s, "=s") {
				s = strings.Split(s, "=s")[0] + "=s0"
			}

			return s
		}
	}

	return ""
}

func (js *Json) String() string {
	if str, ok := (js.Data).(string); ok {
		if str == "" {
			return "None"
		}

		return str
	}

	return ""
}

func (js *Json) Int() int {
	switch num := js.Data.(type) {
	case float32, float64:
		return int(num.(float64))
	case int, int8, int16, int32, int64:
		return int(num.(int64))
	case uint, uint8, uint16, uint32, uint64:
		return int(num.(uint64))
	case json.Number:
		i, err := num.Int64()
		if err == nil {
			return int(i)
		}
	case string:
		i, err := strconv.Atoi(num)
		if err == nil {
			return i
		}
	}

	return 0
}

func (js *Json) Bool() bool {
	if b, ok := (js.Data).(bool); ok {
		return b
	}

	return false
}

func (js *Json) Time() Time {
	if js == (&Json{}) {
		return Time{}
	}

	s := js.String()

	if ts, err := time.Parse(time.RFC3339, s); err == nil {
		return Time(ts)
	} else if ts, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return Time(ts)
	} else if ts, err := time.Parse("2006-01-02", s); err == nil {
		return Time(ts)
	}

	return Time{}
}

func (js *Json) Duration() Duration {
	if js == (&Json{}) {
		return Duration(0)
	}

	var temp string

	if js.HasPrefix("PT") {
		temp = js.Slice(2, -1)
	} else {
		temp = js.String()
	}

	ds, err := time.ParseDuration(strings.ToLower(temp))
	if err != nil {
		return Duration(0)
	}

	return Duration(ds)
}

// for debug
func (js *Json) ToString() string {
	jsonBytes, err := json.Marshal(js)
	if err != nil {
		return ""
	}

	return string(jsonBytes)
}
