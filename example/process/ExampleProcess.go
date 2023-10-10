package process

import (
	. "github.com/Bunny3th/easy-workflow/workflow/engine"
	. "github.com/Bunny3th/easy-workflow/workflow/model"
	"log"
)

//这里定义一个示例流程，而后将流程定义转为Json
func CreateProcessJson() (string, error) {

	//初始节点
	//建议所有的初始节点User都定义为变量"$starter,方便之后的开发管理
	Node1 := Node{NodeID: "Start", NodeName: "请假",
		NodeType: 0, UserIDs: []string{"$starter"},
		NodeEndEvents: []string{"MyEvent_End"},
	}

	//网关,根据请假天数做判断:
	//请假天数>=3,流程转到主管审批
	//请假天数<3,流程结束
	GWConfig_Conditional := HybridGateway{[]Condition{{Expression: "$days>=3", NodeID: "Manager"},
		{Expression: "$days<3", NodeID: "END"}}, []string{}, 0}
	Node2 := Node{NodeID: "GW-Day", NodeName: "请假天数判断",
		NodeType: 2, GWConfig: GWConfig_Conditional,
		PrevNodeIDs: []string{"Start"},
	}

	//主管审批节点
	//注意，这里使用了角色。因为系统无法预先知道角色中存在多少用户，所以必须用StartEvents解析角色，将角色中的用户加到UserIDs中
	//在这个节点中，使用了MyEvent_ResolveRoles做角色解析,MyEvent_Notify做通知
	Node3 := Node{NodeID: "Manager", NodeName: "主管审批",
		NodeType: 1, Roles: []string{"主管"},
		PrevNodeIDs: []string{"GW-Day"},
		NodeStartEvents: []string{"MyEvent_ResolveRoles", "MyEvent_Notify"},
		NodeEndEvents:   []string{"MyEvent_End"},
	}

	//网关,要求“主管审批节点”通过后，并行进入到“人事审批”以及“副总审批”节点
	GW_Parallel := HybridGateway{nil, []string{"HR", "DeputyBoss"}, 0}
	Node4 := Node{NodeID: "GW-Parallel", NodeName: "并行网关",
		NodeType: 2, GWConfig: GW_Parallel,
		PrevNodeIDs: []string{"Manager"},
	}

	//人事审批任务节点，同样使用MyEvent_ResolveRoles做角色解析,MyEvent_Notify做通知
	Node5 := Node{NodeID: "HR", NodeName: "人事审批",
		NodeType: 1, Roles: []string{"人事经理"},
		PrevNodeIDs: []string{"GW-Parallel"},
		NodeStartEvents: []string{"MyEvent_ResolveRoles", "MyEvent_Notify"},
		NodeEndEvents:   []string{"MyEvent_End"},
	}

	//副总审批任务节点
	//注意，IsCosigned=1说明这是一个会签节点，全部通过才算通过，一人驳回即算驳回
	//另外，这里使用了Task结束事件，请务必查看一下该事件逻辑
	Node6 := Node{NodeID: "DeputyBoss", NodeName: "副总审批",
		NodeType: 1, Roles: []string{"副总"},
		IsCosigned:  1,
		PrevNodeIDs: []string{"GW-Parallel"},
		NodeStartEvents: []string{"MyEvent_ResolveRoles", "MyEvent_Notify"},
		NodeEndEvents:   []string{"MyEvent_End"},
		TaskFinishEvents: []string{"MyEvent_TaskForceNodePass"},
	}

	//此网关承接上一个NodeID=GW-Parallel的网关
	//WaitForAllPrevNode=1 等于并行网关，必须要上级节点"人事、副总"全部完成才能往下走
	GW_Parallel2 := HybridGateway{nil, []string{"Boss"}, 1}
	Node7 := Node{NodeID: "GW-Parallel2", NodeName: "并行网关",
		NodeType:    2,
		PrevNodeIDs: []string{"HR", "DeputyBoss"},
		GWConfig:    GW_Parallel2,
	}

	//老板审批任务节点
	//这是一个非会签节点：一人通过即通过，一人驳回即驳回
	Node8 := Node{NodeID: "Boss", NodeName: "老板审批",
		NodeType: 1, Roles: []string{"老板"},
		PrevNodeIDs: []string{"GW-Parallel2"},
		NodeStartEvents: []string{"MyEvent_ResolveRoles", "MyEvent_Notify"},
		NodeEndEvents:   []string{"MyEvent_End"},
	}

	//结束节点
	Node9 := Node{NodeID: "END", NodeName: "END",
		NodeType: 3, PrevNodeIDs: []string{"GW-Day", "Boss"},
		NodeStartEvents: []string{"MyEvent_Notify"}}

	//流程是节点的集合，所以要把上面所有的节点放在一个切片中
	var Nodelist []Node
	Nodelist = append(Nodelist, Node1)
	Nodelist = append(Nodelist, Node2)
	Nodelist = append(Nodelist, Node3)
	Nodelist = append(Nodelist, Node4)
	Nodelist = append(Nodelist, Node5)
	Nodelist = append(Nodelist, Node6)
	Nodelist = append(Nodelist, Node7)
	Nodelist = append(Nodelist, Node8)
	Nodelist = append(Nodelist, Node9)

	process:=Process{ProcessName: "员工请假",Source: "办公系统",RevokeEvents: []string{"MyEvent_Revoke"},Nodes: Nodelist}

	//转化为json
	j, err := JSONMarshal(process, false)

	return string(j), err
}

func CreateExampleProcess(){
	//获得示例流程json
	j, err := CreateProcessJson()
	if err != nil {
		log.Fatal(err)
	}

	//保存流程
	id, err := ProcessSave(j,"system")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("流程保存成功，ID：", id)
}