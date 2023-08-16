package test

import (
	. "easy-workflow/pkg/workflow/engine"
	. "easy-workflow/pkg/workflow/event"
	. "easy-workflow/pkg/workflow/model"
	. "easy-workflow/pkg/workflow/util"
	"fmt"
	"reflect"
	"testing"
)

func ProcessToJson() (string, error) {
	Node1 := Node{NodeID: "A", NodeName: "请假",
		NodeType: 0, UserIDs: []string{"$starter"},
	}

	GWB := HybridGateway{[]Condition{{Expression: "$days>=3", NodeID: "C"}, {Expression: "$days<3", NodeID: "END"}}, []string{}, 0}

	Node2 := Node{NodeID: "B", NodeName: "请假天数判断",
		NodeType: 2, GWConfig: GWB,
		PrevNodeIDs: []string{"A"},
	}

	Node3 := Node{NodeID: "C", NodeName: "主管审批",
		NodeType: 1, UserIDs: []string{"$Manager"},
		PrevNodeIDs: []string{"B"},
	}

	GWD := HybridGateway{nil, []string{"E","F"}, 0}

	Node4 := Node{NodeID: "D", NodeName: "并行网关",
		NodeType: 2, GWConfig: GWD,
		PrevNodeIDs: []string{"C"},
	}

	Node5 := Node{NodeID: "E", NodeName: "老板审批",
		NodeType: 1, UserIDs: []string{"$Boss"},
		PrevNodeIDs: []string{"D"},
	}

	Node6 := Node{NodeID: "F", NodeName: "老板娘审批",
		NodeType: 1, UserIDs: []string{"$Boss2"},
		PrevNodeIDs: []string{"D"},
	}

	GWG := HybridGateway{nil, []string{"END"}, 1}
	Node7 := Node{NodeID: "G", NodeName: "并行网关",
		NodeType:    2,
		PrevNodeIDs: []string{"E", "F"},
		GWConfig: GWG,
	}

	NodeE := Node{NodeID: "END", NodeName: "END",
		NodeType: 3, PrevNodeIDs: []string{"B", "G"}}

	var Nodelist []Node
	Nodelist = append(Nodelist, Node1)
	Nodelist = append(Nodelist, Node2)
	Nodelist = append(Nodelist, Node3)
	Nodelist = append(Nodelist, Node4)
	Nodelist = append(Nodelist, Node5)
	Nodelist = append(Nodelist, Node6)
	Nodelist = append(Nodelist, Node7)
	Nodelist = append(Nodelist, NodeE)

	j, err := JSONMarshal(Nodelist, false)

	return string(j), err
}

//测试将流程定义转为json
func Test_ProcessSave(t *testing.T) {
	j, err := ProcessToJson()
	if err != nil {
		t.Fatal(err)
	}

	id, err := ProcessSave("员工请假", string(j), "001", "办公系统")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("流程保存成功，ID：", id)
}



type event func(ProcessInstanceID int, CurrentNode Node, PrevNode Node) error

func Event1 (ProcessInstanceID int, CurrentNode Node, PrevNode Node) error{
	fmt.Println("this is event")
	return nil
}

func Test_Event(T *testing.T){
	//t:=reflect.TypeOf(Event1)
	//
	//t2:=reflect.TypeOf((event)(nil))
	//
	//fmt.Println(t.AssignableTo(t2))


	EventValue := reflect.ValueOf(&Event{})
	EventType := EventValue.Type()
	for i := 0; i < EventType.NumMethod(); i++ {
		fmt.Println("func:", EventType.Method(i).Name)
		m:=EventType.Method(i)
		run:=reflect.TypeOf((EventRun)(nil))

		fmt.Println(m.Type.NumIn())
		fmt.Println(m.Name)

		//fmt.Println(m.Type.In(0).ConvertibleTo(reflect.TypeOf(Node{})  ))
		//fmt.Println(m.Type.In(1).Kind().String())
		//fmt.Println(m.Type.In(2).ConvertibleTo(reflect.TypeOf(Node{})  ))
		//fmt.Println(m.Type.In(3).ConvertibleTo(reflect.TypeOf(Node{})  ))

		//fmt.Println((error)(nil))

		//e:=reflect.TypeOf((error)(nil))

		errtype:=reflect.TypeOf((*error)(nil)).Elem()
		fmt.Println(errtype)

		fmt.Println(m.Type.Out(0).Implements(errtype))
//var e error=errors.New("")
//		fmt.Println(reflect.ValueOf(out).CanConvert(reflect.TypeOf(e)))

//errors.As()




		if m.Type.AssignableTo(run){
        fmt.Println("yes")
		//var args = []reflect.Value{
		//	reflect.ValueOf(&Event{}),
		//	reflect.ValueOf(1), // 这里要加这个，否则报错
		//	reflect.ValueOf(Node{}),
		//	reflect.ValueOf(Node{}),
		//}
		//EventType.Method(i).Func.Call(args)
	}
}

}

