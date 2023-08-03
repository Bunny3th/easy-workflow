package main

import (
	"easy-workflow/pkg/workflow/model/gateway"
	. "easy-workflow/pkg/workflow/model/node"
	."easy-workflow/pkg/workflow/engine"
	"encoding/json"
	"fmt"
)

func main() {

	//Node.NodeID = "start"
	//Node.NodeName = "开始"
	//Node.NodeType = 0
	//Node.Candidates = []model.Candidate{model.Candidate{UserID: 1}}
	//Node.Comment = "开始请假流程"
	//Node.CounterSign = false
	//var GW model.GW
	//GW = &model.ExclusiveGateway{Key: "key",RegularExpression: "dddddd"}
	//Node.GateWay = []model.GW{GW}



	Node1 := Node{NodeID: "A", NodeName: "请假",
		NodeType: 0, UserIDs: []string{"$starter"},
		}

	//var GW gateway.GateWay
	GW := gateway.Gateway{Key: "请假类型判断", RegularExpression: "dddddd", NodeIDs: []string{"C"}}

	Node2 := Node{NodeID: "B", NodeName: "请假类s型判断",
		NodeType: 2, GateWay: GW,
		PrevNodeIDs: []string{"A"},
		}

	Node3 := Node{NodeID: "C", NodeName: "审批",
		NodeType: 1, UserIDs: []string{"002"},
		PrevNodeIDs: []string{"B"},
		}

	NodeE := Node{NodeID: "END" , NodeName: "END",
		NodeType: 3,PrevNodeIDs: []string{"C","D"}}

	var Nodelist []Node
	Nodelist = append(Nodelist, Node1)
	Nodelist = append(Nodelist, Node2)
	Nodelist = append(Nodelist, Node3)
	Nodelist = append(Nodelist, NodeE)

	j, err := json.Marshal(Nodelist)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(j))

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

	//id,err=InstanceInit(1,"Business123",map[string]string{"starter":"U0001"})
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println("流程实例ID:",id)
	ProcessNode(1,Node{})

}
