package node

import (
	. "easy-workflow/pkg/workflow/model/gateway"
)

type NodeType int

const (
	Root    NodeType = 0 //开始节点
	General NodeType = 1 //任务节点,指的是需要人完成的节点
	GateWay NodeType = 2 //参考activiti的网关.目前只实现了排他网关
	End     NodeType = 3 //结束节点,结束节点不需要人参与，到了此节点，则流程实例完成
)

/*思考
一、为什么需要一个结束节点:
1、强制流程必须有一个结束节点，可以防止流程卡死
2、不在任务节点上设置结束标记，因为这样可能在分支较多的时候，出现N个带结束标记的任务节点。可能由于疏忽，改分支流转的最后一个任务节点没有加结束标记，则流程卡死
二、为什么必须要一个开始节点
开始节点意味着流程实例的创建。
BPMBN 2.0标准是必须要有开始和结束节点的。但是思考下来，似乎在这里并不需要一个空的开始节点，可以把开始和任务节点结合。

结束节点这里有问题，如果多个任务节点都指向结束节点，那么结束节点的prev就会有很多个，
结束节点应该不用定义任何候选人，只要一个任务完成，下一个节点是结束节点，那么就应该结束
*/

type Node struct {
	NodeID      string   //节点名称
	NodeName    string   //节点名字
	NodeType    NodeType //节点类型
	PrevNodeIDs []string //上级节点(任何任务节点只能有一个上级节点;结束节点可以有多个上级节点)
	NextNodeIDs []string //下级节点(节点可以有N个直接下级节点)  是否需要下级节点？感觉可以不用。需要，因为生成task时候需要知道下级是谁，或者可以通过计算得知
	UserIDs     []string //节点处理人数组
	//Role        []string //节点处理角色数组。注意，一旦使用角色，则该节点默认不能会签。因为系统无法预先知道角色中存在多少用户。除非通过事件修改.暂时不用
	GateWay Gateway //网关。只有在节点类型为GateWay的情况下此字段才会有值
	//Comment   string            //备注应该是运行时task输入
	IsCosigned int8     //是否会签  会签的情况下，需要所有人通过才能进行下一节点，只要有一人反对，则退回上一节点
	PreEvents  []string //前置事件
	ExitEvents []string //退出事件
	//Variables   map[string]string //传入的变量  这个不应该在node中定义，应该是在传入参数中获取
}
