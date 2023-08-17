package event

import (
	. "easy-workflow/workflow/model"
	. "easy-workflow/workflow/util"
	"fmt"
	"log"
	"reflect"
)

type Method struct {
	S interface{} //method所在的struct，这是函数执行的第一个参数
	M reflect.Method
}

//事件池，所有的事件都注册在这里
//var EventPool = make(map[string]reflect.Method)
var EventPool = make(map[string]Method)

//注册一个struct中的所有func
//注意，func签名必须是func(struct *interface{}, ProcessInstanceID int, CurrentNode *Node, PrevNode Node) error
func RegisterEvents(Struct any) {
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

		if m.Type.In(2).ConvertibleTo(reflect.TypeOf(&Node{})) != true {
			log.Printf("warning:事件方法 %s 参数2不是*Node类型,此函数不会被导入", m.Name)
			continue
		}

		if m.Type.In(3).ConvertibleTo(reflect.TypeOf(Node{})) != true {
			log.Printf("warning:事件方法 %s 参数3不是Node类型,此函数不会被导入", m.Name)
			continue
		}

		if !TypeIsError(m.Type.Out(0)) {
			log.Printf("warning:事件方法 %s 返回参数不是error类型,此函数不会被导入", m.Name)
			continue
		}

		var method = Method{Struct, m}

		EventPool[m.Name] = method
	}
}

//检查流程节点中事件是否已经被注册
func CheckIfEventImported(ProcessNode Node) error {
	//首先合并节点的PreEvents和ExitEvents
	var events []string
	events = append(events, ProcessNode.PreEvents...)
	events = append(events, ProcessNode.ExitEvents...)
	//判断该节点中是否所有事件都已经被注册
	for _, event := range events {
		if _, ok := EventPool[event]; !ok {
			return fmt.Errorf("事件%s尚未导入", event)
		}
	}

	return nil
}

//运行事件
func RunEvent(EventName string, ProcessInstanceID int, CurrentNode *Node, PrevNode Node) error {
	log.Printf("正在处理节点[%s]中事件[%s]", CurrentNode.NodeName, EventName)
	//判断时候可以在事件池中获取事件
	event, ok := EventPool[EventName]
	if !ok {
		return fmt.Errorf("事件%s未注册", EventName)
	}

	//拼装参数
	arg := []reflect.Value{
		reflect.ValueOf(event.S),
		reflect.ValueOf(ProcessInstanceID),
		reflect.ValueOf(CurrentNode),
		reflect.ValueOf(PrevNode),
	}

	//运行func
	result := event.M.Func.Call(arg)

	//判断第一个返回参数是否为nil
	if !result[0].IsNil() {

		return fmt.Errorf("节点[%s]事件[%s]执行出错:%v", CurrentNode.NodeName, event.M.Name, result[0])
	}

	return nil
}
