package event

import (
	. "easy-workflow/workflow/model"
	"fmt"
	"log"
	"reflect"
)

type EventRun func(event *interface{}, ProcessInstanceID int, CurrentNode Node, PrevNode Node) error

type Event struct{}

func (e *Event) MyEvent(ProcessInstanceID int, CurrentNode Node, PrevNode Node) error {
	fmt.Println("fucking shit!!!!!")
	return nil
}


var EventCache = make(map[string]reflect.Method)

func ImportEvents(Struct *any) {
	StructValue := reflect.ValueOf(Struct)
	StructType := StructValue.Type()

	for i := 0; i < StructType.NumMethod(); i++ {
		m := StructType.Method(i)
		//自定义函数必须是4个参数，参数0：*struct{} 1:int 2:Node 3:Node
		if m.Type.NumIn() != 4 || m.Type.NumOut() != 1 {
			log.Printf("warning:事件方法 %s 入参、出参数量不匹配,此函数不会被导入", m.Name)
			continue
		}

		if m.Type.In(1).Kind().String() != "int" {
			log.Printf("warning:事件方法 %s 参数1不是int类型,此函数不会被导入", m.Name)
			continue
		}

		if m.Type.In(2).ConvertibleTo(reflect.TypeOf(Node{})) != true {
			log.Printf("warning:事件方法 %s 参数2不是Node类型,此函数不会被导入", m.Name)
			continue
		}

		if m.Type.In(3).ConvertibleTo(reflect.TypeOf(Node{})) != true {
			log.Printf("warning:事件方法 %s 参数3不是Node类型,此函数不会被导入", m.Name)
			continue
		}

		errType := reflect.TypeOf((*error)(nil)).Elem()
		if m.Type.Out(0).Implements(reflect.TypeOf(errType)) {
			log.Printf("warning:事件方法 %s 返回参数不是error类型,此函数不会被导入", m.Name)
			continue
		}

		EventCache[m.Name] = m
	}
}

