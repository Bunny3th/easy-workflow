package main

import (
	"bytes"
	//. "easy-workflow/pkg/workflow/engine"
	. "easy-workflow/pkg/workflow/model"
	"encoding/json"
	"fmt"
)

func JSONMarshal(t interface{}, escapeHtml bool) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(escapeHtml)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}


func main() {

	Node1 := Node{NodeID: "A", NodeName: "请假",
		NodeType: 0, UserIDs: []string{"$starter"},
	}

	//var GW gateway.GateWay
	GW := ExclusiveGateway{[]Condition{{Expression: "$days>=3", NodeID: "C"}, {Expression: "$days<3", NodeID: "END"}}}

	Node2 := Node{NodeID: "B", NodeName: "请假天数判断",
		NodeType: 2, GWConfig: GW,
		PrevNodeIDs: []string{"A"},
	}

	Node3 := Node{NodeID: "C", NodeName: "主管审批",
		NodeType: 1, UserIDs: []string{"$Manager"},
		PrevNodeIDs: []string{"B"},
	}

	Node4 := Node{NodeID: "D", NodeName: "老板审批",
		NodeType: 1, UserIDs: []string{"$Boss"},
		PrevNodeIDs: []string{"C"},
	}

	NodeE := Node{NodeID: "END", NodeName: "END",
		NodeType: 3, PrevNodeIDs: []string{"D", "B"}}

	var Nodelist []Node
	Nodelist = append(Nodelist, Node1)
	Nodelist = append(Nodelist, Node2)
	Nodelist = append(Nodelist, Node3)
	Nodelist = append(Nodelist, Node4)
	Nodelist = append(Nodelist, NodeE)

	j,err:= JSONMarshal(Nodelist,false)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print(string(j))


	//j, err := json.Marshal(Nodelist)



	//id,err := ProcessSave("员工请假", string(j), "001", "SYSA")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println("流程保存成功，ID：",id)

	//ID,err:=GetProcessID("员工请假","SYSA")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println("ProcessID",ID)

	//nodes,err:=GetProcessDefine(4)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Printf("%+v",nodes)

	//id, err := InstanceStart(1, "Business123", "请假啦", map[string]string{"starter": "U0001", "Manager": "U0002", "days": "5"})
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println("流程实例ID:", id)

	//应该在pass的时候直接处理下一个
	//TaskPass(4,"审批通过","")
	//TaskReject(2,"审批通过","")
}
