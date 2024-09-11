package engine

import (
	"fmt"
	. "github.com/Bunny3th/easy-workflow/workflow/model"
	"reflect"
)

type method struct {
	S interface{}    //method所在的struct，这是函数执行的第一个参数
	M reflect.Method //方法
}

//事件池，所有的事件都会在流程引擎启动的时候注册到这里
var EventPool = make(map[string]method)

//事件出错，则可能导致流程无法运行下去,在这里添加选项，是否忽略事件出错，让流程继续
var IgnoreEventError bool

//注册一个struct中的所有func
//注意,此时不会验证事件方法参数是否正确,因为此时不知道事件到底是“节点事件”还是“流程事件”
func RegisterEvents(Struct any) {
	StructValue := reflect.ValueOf(Struct)
	StructType := StructValue.Type()

	for i := 0; i < StructType.NumMethod(); i++ {
		m := StructType.Method(i)
		var method = method{Struct, m}
		EventPool[m.Name] = method
	}
}

//验证流程事件(目前只有流程撤销事件)参数是否正确
//流程撤销事件  func签名必须是func(struct *interface{}, ProcessInstanceID int,RevokeUserID string) error
func verifyProcEventParameters(m reflect.Method) error {
	//自定义函数必须是3个参数，参数0：*struct{} 1:int 2:String
	if m.Type.NumIn() != 3 || m.Type.NumOut() != 1 {
		return fmt.Errorf("warning:事件方法 %s 入参、出参数量不匹配,此函数无法运行", m.Name)
	}

	if m.Type.In(1).Kind().String() != "int" {
		return fmt.Errorf("warning:事件方法 %s 参数1不是int类型,此函数无法运行", m.Name)
	}

	if m.Type.In(2).Kind().String() != "string" {
		return fmt.Errorf("warning:事件方法 %s 参数2不是string类型,此函数无法运行", m.Name)
	}

	if !TypeIsError(m.Type.Out(0)) {
		return fmt.Errorf("warning:事件方法 %s 返回参数不是error类型,此函数无法运行", m.Name)
	}
	return nil
}

//验证节点事件(1、节点开始  2、节点结束 3、任务结束)参数是否正确
//1、节点开始、结束事件     func签名必须是func(struct *interface{}, ProcessInstanceID int, CurrentNode *Node, PrevNode Node) error
//2、任务完成事件          func签名必须是func(struct *interface{}, TaskID int, CurrentNode *Node, PrevNode Node) error
func verifyNodeEventParameters(m reflect.Method) error {
	//自定义函数必须是4个参数，参数0：*struct{} 1:int 2:Node 3:Node
	if m.Type.NumIn() != 4 || m.Type.NumOut() != 1 {
		return fmt.Errorf("warning:事件方法 %s 入参、出参数量不匹配,此函数无法运行", m.Name)
	}

	if m.Type.In(1).Kind().String() != "int" {
		return fmt.Errorf("warning:事件方法 %s 参数1不是int类型,此函数无法运行", m.Name)
	}

	if m.Type.In(2).ConvertibleTo(reflect.TypeOf(&Node{})) != true {
		return fmt.Errorf("warning:事件方法 %s 参数2不是*Node类型,此函数无法运行", m.Name)
	}

	if m.Type.In(3).ConvertibleTo(reflect.TypeOf(Node{})) != true {
		return fmt.Errorf("warning:事件方法 %s 参数3不是Node类型,此函数无法运行", m.Name)
	}

	if !TypeIsError(m.Type.Out(0)) {
		return fmt.Errorf("warning:事件方法 %s 返回参数不是error类型,此函数无法运行", m.Name)
	}
	return nil
}

//检查流程:
//1、是否注册
//2、参数是否正确
func VerifyEvents(ProcessID int, Nodes ProcNodes) error {
	//获取流程定义
	process, err := GetProcessDefine(ProcessID)
	if err != nil {
		return err
	}

	//验证流程事件(目前只有撤销事件)
	for _, event := range process.RevokeEvents {
		if e, ok := EventPool[event]; !ok {
			return fmt.Errorf("事件%s尚未导入", event)
		} else {
			if err := verifyProcEventParameters(e.M); err != nil {
				return err
			}
		}
	}

	//各个节点中开始、结束事件 and 任务完成事件,先放入一个数组
	var nodeEvents []string
	for _, node := range Nodes {
		nodeEvents = append(nodeEvents, node.NodeStartEvents...)
		nodeEvents = append(nodeEvents, node.NodeEndEvents...)
		nodeEvents = append(nodeEvents, node.TaskFinishEvents...)
	}

	//各个节点中事件可能有重复的，需做去重
	nodeEventsSet := MakeUnique(nodeEvents)

	//验证节点事件
	for _, event := range nodeEventsSet {
		if e, ok := EventPool[event]; !ok {
			return fmt.Errorf("事件%s尚未导入", event)
		} else {
			if err := verifyNodeEventParameters(e.M); err != nil {
				return err
			}
		}
	}

	return nil
}

//运行节点事件(1、节点开始  2、节点结束 3、任务结束)
func RunNodeEvents(EventNames []string, ID int, CurrentNode *Node, PrevNode Node) error {
	for _, e := range EventNames {
		//log.Printf("正在处理节点[%s]中事件[%s]", CurrentNode.NodeName, e)
		//判断是否可以在事件池中获取事件
		event, ok := EventPool[e]
		if !ok {
			return fmt.Errorf("事件%s未注册", e)
		}

		//拼装参数
		arg := []reflect.Value{
			reflect.ValueOf(event.S),
			reflect.ValueOf(ID),
			reflect.ValueOf(CurrentNode),
			reflect.ValueOf(PrevNode),
		}

		//运行func
		result := event.M.Func.Call(arg)

		//如果选项IgnoreEventError为false,则说明需要验证事件是否出错
		if IgnoreEventError == false {
			//判断第一个返回参数是否为nil,若不是，则说明事件出错
			if !result[0].IsNil() {
				return fmt.Errorf("节点[%s]事件[%s]执行出错:%v", CurrentNode.NodeName, event.M.Name, result[0])
			}
		}
	}

	return nil
}

////运行流程事件(目前只有撤销事件)
func RunProcEvents(EventNames []string, ProcessInstanceID int, RevokeUserID string) error {
	for _, e := range EventNames {
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
			reflect.ValueOf(RevokeUserID),
		}

		//获取流程名
		processName, err := GetProcessNameByInstanceID(ProcessInstanceID)
		if err != nil {
			return err
		}

		//运行func
		result := event.M.Func.Call(arg)

		//如果选项IgnoreEventError为false,则说明需要验证事件是否出错
		if IgnoreEventError == false {
			//判断第一个返回参数是否为nil,若不是，则说明事件出错
			if !result[0].IsNil() {
				return fmt.Errorf("流程[%s]撤销事件[%s]执行出错:%v", processName, event.M.Name, result[0])
			}
		}
	}

	return nil
}
