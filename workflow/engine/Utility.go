package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strings"
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

func TypeIsError(Type reflect.Type) bool {
	//只要实现 Error() string方法的，就认为是实现了error接口
	//所以，要判断type中是否有一个方法名叫Error，无传入参数，输出string

	//如果都没有方法，自然不可能实现error
	if Type.NumMethod() >= 1 {
		for i := 0; i < Type.NumMethod(); i++ {
			if Type.Method(i).Name == "Error" {
				//是否无传入参数
				if Type.Method(i).Type.NumIn() != 0 {
					return false
				}

				//是否只有一个输出参数
				if Type.Method(i).Type.NumOut() != 1 {
					return false
				}

				//输出参数是否是string
				if Type.Method(i).Type.Out(0).Kind().String() != "string" {
					return false
				}

				return true
			}
		}
	}

	return false
}

//这是一个偷懒取巧的办法:利用数据库计算网关中的条件表达式
func ExpressionEvaluator(expression string) (bool, error) {
	//首先通过正则表达式判断是否有SQL注入风险
	pattern := regexp.MustCompile("delete|truncate|insert|drop|create|select|update|set|from|grant|call|execute")
	match := pattern.FindString(strings.ToLower(expression))
	if match != "" {
		return false, errors.New("表达式中包含危险词,可能造成SQL注入!")
	}

	sql := "SELECT " + expression
	var ok bool
	_, err := ExecSQL(sql, &ok)
	if err != nil {
		return false, err
	}
	return ok, nil
}

//利用Map，对数组/切片数据做去重处理
func MakeUnique(List ...[]string) []string {
	set := make(map[string]string)
	var unique []string

	for _, list := range List {
		for _, item := range list {
			set[item] = ""
		}
	}

	for k, _ := range set {
		unique = append(unique, k)
	}

	return unique
}

// removeAllElements 删除切片中所有匹配的元素 2024.11.13 add by yujf
func RemoveAllElements(slice []string, value string) []string {
	result := []string{}
	for _, v := range slice {
		if v != value {
			result = append(result, v)
		}
	}
	return result
}
