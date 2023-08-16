package util

import (
	"bytes"
	"encoding/json"
)

//将json字符串转为struct
func Json2Struct(j string, s any) error {
	return json.Unmarshal([]byte(j), s)
}

//json.Marshal()函数默认用HTMLEscape进行编码，它将替换“＜”、“＞”、“&”、U+2028和U+2029，
//并将其转义为“\u003c”、“\u003e”、“\u0026”、“\ u2028”和“\u2029”
//所以在这里做处理，判断是否开启转义
func JSONMarshal(t interface{}, escapeHtml bool) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(escapeHtml)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}