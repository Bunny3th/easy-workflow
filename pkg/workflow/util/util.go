package util

import (
	"encoding/json"
)

//将json字符串转为struct
func Json2Struct(j string, s any) error {
	return json.Unmarshal([]byte(j), s)
}
