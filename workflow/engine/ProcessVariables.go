package engine

import (
	"easy-workflow/workflow/dao"
	"encoding/json"
	"errors"
	"strings"
)

//判断传入字符串是否是变量(是否以$开头)
func IsVariable(Key string) bool {
	if strings.HasPrefix(Key, "$") {
		return true
	}
	return false
}

//去掉变量前缀"$"
func RemovePrefix(variable string) string {
	return strings.Replace(variable, "$", "", -1)
}

//从proc_inst_variable表中查找变量，若有则返回变量值,若无则返回false
func SetVariable(ProcessInstanceID int, variable string) (string, bool, error) {
	Key := RemovePrefix(variable)
	type result struct {
		Value string
	}
	var r result
	if _, err := dao.ExecSQL("SELECT `value` FROM `proc_inst_variable` "+
		"WHERE `proc_inst_id`=? AND `key`=? LIMIT 1", &r, ProcessInstanceID, Key); err == nil {

		//判断是否有匹配的值
		exists := false
		if r.Value != "" {
			exists = true
		}

		return r.Value, exists, nil
	} else {
		return "", false, err
	}
}

//将变量map生成kv对形式的json字符串，以便存入数据库
//此方法暂时无用
func VariablesMap2Json(Variables map[string]string) (string, error) {
	type kv struct {
		Key   string
		Value string
	}
	var kvs []kv
	for k, v := range Variables {
		kvs = append(kvs, kv{Key: k, Value: v})
	}
	j, err := json.Marshal(kvs)
	if err != nil {
		return "", err
	}
	return string(j), nil
}


//解析变量,获取并设置其value,返回map(注意，如果不是变量，则原样存储在map中)
func ResolveVariables(ProcessInstanceID int, Variables []string) (map[string]string, error) {
	result:=make(map[string]string)
	for _, v := range Variables {
		if IsVariable(v) {
			value, ok, err := SetVariable(ProcessInstanceID, v)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, errors.New("无法匹配变量:" + v)
			}
			result[v] = value
		} else {
			result[v] =v
		}
	}
	return result, nil
}

