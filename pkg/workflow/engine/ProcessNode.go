package engine

import (
	"easy-workflow/pkg/dao"
	. "easy-workflow/pkg/workflow/model/node"
	"errors"
	"regexp"
	"strings"
)

//处理节点,如：生成task、进行条件判断、处理结束节点等
func ProcessNode(ProcessInstanceID int, CurrentNode Node, PrevNode Node) error {
	//这里处理前置事件
	//do something

	//这里不需要处理开始节点，因为开始节点是特殊的，一个流程只有一个开始节点
	//开始节点也需要处理，因为开始节点可能因为驳回而重新回到开始节点
	if CurrentNode.NodeType == RootNode {
		err := TaskNodeHandle(ProcessInstanceID, CurrentNode, PrevNode)
		if err != nil {
			return err
		}
	}


	if CurrentNode.NodeType == GateWayNode {
		err := GateWayNodeHandle(ProcessInstanceID, CurrentNode, PrevNode)
		if err != nil {
			return err
		}
	}

	if CurrentNode.NodeType == TaskNode {
		err := TaskNodeHandle(ProcessInstanceID, CurrentNode, PrevNode)
		if err != nil {
			return err
		}
	}

	if CurrentNode.NodeType == EndNode {
		err := EndNodeHandle(ProcessInstanceID)
		if err != nil {
			return err
		}
	}

	return nil

	//这里处理退出事件
	//do something

}

//开始节点处理
func StartNodeHandle(ProcessInstanceID int, StartNode Node, Comment string, VariableJson string) error {
	if StartNode.NodeType != RootNode {
		return errors.New("不是开始节点，无法处理节点:" + StartNode.NodeName)
	}

	//开始节点只需要一个处理人
	//思考：如果开始节点有多个处理人，则可能进入会签状态，可能造成流程都无法正常开始
	//所以，开始节点只能有一个处理人，且默认就是非会签节点
	if len(StartNode.UserIDs) < 1 {
		return errors.New("未指定处理人，无法处理节点:" + StartNode.NodeName)
	}

	//匹配节点用户变量,开始节点只能有一个用户发起,所以不管多少用户，只要传第一个
	users, err := ResolveVariables(ProcessInstanceID, StartNode.UserIDs[0:1])
	if err != nil {
		return err
	}

	//生成一条Task
	taskids, err := CreateTask(ProcessInstanceID, StartNode.NodeID, "", users)
	if err != nil {
		return err
	}

	//完成task,并获取下一步NodeID
	err = TaskPass(taskids[0], Comment, VariableJson)
	if err != nil {
		return err
	}

	return nil

}

//结束节点处理
func EndNodeHandle(ProcessInstanceID int) error {
	_,err:=dao.ExecSQL("call sp_proc_inst_end(?)",nil,ProcessInstanceID)
	return err
}

//任务节点处理
func TaskNodeHandle(ProcessInstanceID int, CurrentNode Node, PrevNode Node) error {
	//匹配节点用户变量
	users, err := ResolveVariables(ProcessInstanceID, CurrentNode.UserIDs)
	if err != nil {
		return err
	}

	//生成Task
	_, err = CreateTask(ProcessInstanceID, CurrentNode.NodeID, PrevNode.NodeID, users)
	if err != nil {
		return err
	}

	return nil

}

//思考一个问题，如果task1-gw1-gw2-task2，那么gw2处理的时候，prev就是gw1，之后task2执行的时候，prev是哪个？
//最简单的就是规定不能连续gw节点
//这里应该没有问题了，因为:
//1、只有任务节点才能开启一个gw
//2、在这里已经是直接把任务节点作为PrevTaskNode原样传入下一个ProcessNode了
func GateWayNodeHandle(ProcessInstanceID int, CurrentNode Node, PrevTaskNode Node) error {
	var NodeIDs []string

	for _, c := range CurrentNode.GWConfig.Conditions {

		//正则表达式，匹配以$开头的字母、数字、下划线
		reg := regexp.MustCompile(`[$]\w+`)
		//获取表达式中所有的变量
		variables := reg.FindAllString(c.Expression, -1)

		expression := c.Expression
		//替换变量
		for _, v := range variables {
			//获取变量对应的value
			values, err := ResolveVariables(ProcessInstanceID, variables)
			if err != nil {
				return err
			}
			expression = strings.Replace(expression, v, values[0], -1)
		}

		//计算表达式，如果成功，则将节点添加到下一个节点组中
		var ok bool
		_, err := dao.ExecSQL("call sp_expression_evaluator(?)", &ok, expression)
		if err != nil {
			return err
		}
		if ok {
			NodeIDs = append(NodeIDs, c.NodeID)
		}
	}

	for _, nodeid := range NodeIDs {
		NextNode, ok, err := GetInstanceNode(ProcessInstanceID, nodeid)
		if err != nil {
			return err
		}

		//节点处理
		if ok {
			ProcessNode(ProcessInstanceID, NextNode, PrevTaskNode)
		}
	}

	return nil

}


//获取流程实例中某个Node
func GetInstanceNode(ProcessInstanceID int, NodeID string) (Node, bool, error) {
	ProcID, err := GetProcessIDByInstanceID(ProcessInstanceID)
	if err != nil {
		return Node{}, false, err
	}

	//从Cache中获得节点列表，取出下一个节点
	Nodes, err := GetProcCache(ProcID)
	if err != nil {
		return Node{}, false, err
	}
	Node, ok := Nodes[NodeID]

	return Node, ok, nil
}
