package test

import (
	. "easy-workflow/pkg/workflow/model"
	. "easy-workflow/pkg/workflow/util"
	"fmt"
	"testing"
)

//测试将流程定义转为json
func Test_ProcessToJson(t *testing.T) {
	Node1 := Node{NodeID: "A", NodeName: "请假",
		NodeType: 0, UserIDs: []string{"$starter"},
	}

	GW := HybridGateway{[]Condition{{Expression: "$days>=3", NodeID: "C"}, {Expression: "$days<3", NodeID: "END"}}, []string{}, 0}

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

	j, err := JSONMarshal(Nodelist, false)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Print(string(j))
}
