package main

import (
	. "easy-workflow/pkg/workflow/model/node"
	"easy-workflow/pkg/workflow/model/gateway"
	. "easy-workflow/pkg/workflow/engine/process"
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
		NodeType: 0, UserIDs: []string{"001"},
		//NextNodeIDs: []string{"B"},
		IsCosigned:  0}

	//var GW gateway.GateWay
	GW := gateway.Gateway{Key: "请假类型判断", RegularExpression: "dddddd", NodeIDs: []string{"C"}}

	Node2 := Node{NodeID: "B", NodeName: "请假类型判断",
		NodeType: 2, GateWay: GW,
		PrevNodeID: "A",
		IsCosigned: 0}

	Node3 := Node{NodeID: "C", NodeName: "审批",
		NodeType: 3, UserIDs: []string{"002"},
		PrevNodeID: "B",
		IsCosigned: 0}

	var Nodes []Node
	Nodes = append(Nodes, Node1)
	Nodes = append(Nodes, Node2)
	Nodes = append(Nodes, Node3)

	j, err := json.Marshal(Nodes)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(j))

	//err = ProcessSave("员工请假", string(j), "001", "SYSA")
	//if err != nil {
	//	fmt.Println(err)
	//}

	ID,err:=GetProcessID("员工请假","SYSA")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("ProcessID",ID)
}
