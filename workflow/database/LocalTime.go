package database

import (
	"database/sql/driver"
	"fmt"
	"time"
)

//gorm中，定义数据表datetime字段类型为time.time时,查询返回格式类似2023-09-19T14:41:28+08:00
//这种格式对人不友好，亦对前端处理不友好(js时间处理函数较弱)
//故使用自定义类型，对时间格式做格式化处理
type LocalTime time.Time

var LTime *LocalTime

func (t *LocalTime) MarshalJSON() ([]byte, error) {
	tTime := time.Time(*t)
	return []byte(fmt.Sprintf("\"%v\"", tTime.Format("2006-01-02 15:04:05"))), nil
}
func (t LocalTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	tlt := time.Time(t)
	//判断给定时间是否和默认零时间的时间戳相同
	if tlt.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return tlt, nil
}
func (t *LocalTime) Scan(v interface{}) error {
	if value, ok := v.(time.Time); ok {
		*t = LocalTime(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

func (t *LocalTime) Now() LocalTime {
	return LocalTime(time.Now())
}

func (t *LocalTime) String() string {
	if t == nil {
		return ""
	}

	return time.Time(*t).Format("2006-01-02 15:04:05")
}
