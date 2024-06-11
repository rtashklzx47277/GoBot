package tools

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"
)

type Json struct {
	Data any `json:"data"`
}

func ToJson(data io.Reader) (*Json, error) {
	var js Json

	err := json.NewDecoder(data).Decode(&js.Data)

	if err != nil {
		return &Json{}, err
	}

	return &js, nil
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

func (js *Json) Split(sep string) []string {
	if str, ok := (js.Data).(string); ok {
		return strings.Split(str, sep)
	}

	return []string{}
}

func (js *Json) Image() string {
	for _, size := range []string{"maxres", "standard", "high", "medium", "default"} {
		if js.Exist(size) {
			return strings.Replace(js.Get(size).Get("url").Split("=s")[0], "_live.", ".", 1)
		}
	}

	return DefaultImage
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

	ts, err := time.Parse(time.RFC3339, strings.Replace(js.String(), "+", "Z", 1))
	if err != nil {
		return Time{}
	}

	return Time(ts)
}

func (js *Json) Duration() Duration {
	if js == (&Json{}) {
		return Duration(0)
	}

	var temp string

	if strings.HasPrefix(js.String(), "PT") {
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
