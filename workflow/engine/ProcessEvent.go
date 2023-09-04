package engine

import (
	"fmt"
	. "github.com/Bunny3th/easy-workflow/workflow/model"
	. "github.com/Bunny3th/easy-workflow/workflow/util"
	"log"
	"reflect"
)

type method struct {
	S interface{} //method所在的struct，这是函数执行的第一个参数
	M reflect.Method  //方法
}

//事件池，所有的事件都会在流程引擎启动的时候注册到这里
var EventPool = make(map[string]method)

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

		var method = method{Struct, m}

		EventPool[m.Name] = method
	}
}

//检查流程节点中事件是否已经被注册
func CheckIfEventRegistered(ProcessNode Node) error {
	//首先合并节点的开始和结束事件
	var events []string
	events = append(events, ProcessNode.StartEvents...)
	events = append(events, ProcessNode.EndEvents...)
	//判断该节点中是否所有事件都已经被注册
	for _, event := range events {
		if _, ok := EventPool[event]; !ok {
			return fmt.Errorf("事件%s尚未导入", event)
		}
	}

	return nil
}

//运行事件
func RunEvents(EventNames []string, ProcessInstanceID int, CurrentNode *Node, PrevNode Node) error {
	for _,e:=range EventNames {
		//log.Printf("正在处理节点[%s]中事件[%s]", CurrentNode.NodeName, e)
		//判断是否可以在事件池中获取事件
		event, ok := EventPool[e]
		if !ok {
			return fmt.Errorf("事件%s未注册", e)
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
	}
	return nil
}

