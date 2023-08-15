package engine

import (
	"easy-workflow/pkg/dao"
	. "easy-workflow/pkg/workflow/model"
	"errors"
	"fmt"
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

	//匹配节点用户变量,开始节点只能有一个用户发起,所以不管多少用户，只要第一个
	userid := StartNode.UserIDs[0:1][0]
	users, err := ResolveVariables(ProcessInstanceID, []string{userid})
	if err != nil {
		return err
	}

	//生成一条Task
	taskids, err := CreateTask(ProcessInstanceID, StartNode.NodeID, "", []string{users[userid]})
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
	_, err := dao.ExecSQL("call sp_proc_inst_end(?,?)", nil, ProcessInstanceID,1)
	return err
}

//任务节点处理
func TaskNodeHandle(ProcessInstanceID int, CurrentNode Node, PrevNode Node) error {
	//匹配节点用户变量
	kv, err := ResolveVariables(ProcessInstanceID, CurrentNode.UserIDs)
	if err != nil {
		return err
	}

	//生成用户数组
	var users []string
	for _, v := range kv {
		users = append(users, v)
	}

	//生成Task
	_, err = CreateTask(ProcessInstanceID, CurrentNode.NodeID, PrevNode.NodeID, users)
	if err != nil {
		return err
	}

	return nil

}

//GateWay节点处理
func GateWayNodeHandle(ProcessInstanceID int, CurrentNode Node, PrevTaskNode Node) error {
	//--------------------首先，混合节点需要确认所有的上级节点都处理完，才能做下一步--------------------
	var totalFinished int                          //所有已完成的上级节点
	totalPrevNodes := len(CurrentNode.PrevNodeIDs) //所有上级节点
	for _, nodeID := range CurrentNode.PrevNodeIDs {
		finished, err := IfInstanceNodeIsFinish(ProcessInstanceID, nodeID)
		if err != nil {
			return err
		}
		if finished {
			totalFinished++
		}
	}

	//如果是并行网关模式，还有尚未完成的上级节点，则退出
	if CurrentNode.GWConfig.WaitForAllPrevNode == 1 && totalPrevNodes != totalFinished {
		return nil
	}

	//如果是包含网关模式,连一个已完成的上级节点都没有，则退出
	if CurrentNode.GWConfig.WaitForAllPrevNode == 0 && totalFinished < 1 {
		return nil
	}

	//----------------------------计算条件----------------------------
	var conditionNodeIDs []string //condition指定的下级Node
	//一个GW节点可以有多个condition,所以要遍历
	for _, c := range CurrentNode.GWConfig.Conditions {
		//正则表达式，匹配以$开头的字母、数字、下划线
		reg := regexp.MustCompile(`[$]\w+`)
		//获取表达式中所有的变量
		variables := reg.FindAllString(c.Expression, -1)

		//替换表达式中的变量为值
		expression := c.Expression
		//获取变量对应的value
		kv, err := ResolveVariables(ProcessInstanceID, variables)
		if err != nil {
			return err
		}
		for k, v := range kv {
			expression = strings.Replace(expression, k, v, -1)
		}

		//计算表达式，如果成功，则将节点添加到下一级节点组中
		var ok bool
		_, err = dao.ExecSQL("call sp_expression_evaluator(?)", &ok, expression)
		if err != nil {
			return err
		}
		if ok {
			conditionNodeIDs = append(conditionNodeIDs, c.NodeID)
		}
	}

	//-------将conditionNodeIDs和InevitableNodes中的值一起放入nextNodeIDs，这是真正要处理的节点ID-------
	var nextNodeIDs = make(map[string]string) //这里用map主要是为了去重(节点ID如果重复，意味着一个节点要做N次处理，这是灾难)
	for _, v := range conditionNodeIDs {
		nextNodeIDs[v] = ""
	}
	for _, v := range CurrentNode.GWConfig.InevitableNodes {
		nextNodeIDs[v] = ""
	}

	//------------------------------对下级节点进行处理------------------------------
	for nodeID, _ := range nextNodeIDs {
		NextNode, err := GetInstanceNode(ProcessInstanceID, nodeID)
		if err != nil {
			return err
		}
		/*
			思考一个问题，ProcessNod函数的形参PrevNode应该传什么？
			如果传当前处理的GW节点本身，则要思考以下情况：
			节点定义是task1-gw1-gw2-task2，如果在gw1处理的最后，ProcessNode的PrevNode传gw1本身，那么task2就永远找不到task1了
			所以，在处理gw节点时,ProcessNod函数的形参PrevNode不能传gw本身，而是要传gw的上一节点，因为：
			1、只有任务节点才能开启一个gw
			2、直接把任务节点作为PrevTaskNode传入，就算下一个节点还是gw，重复此行为，之后的task节点还是可以获得上一个task节点
		*/
		err=ProcessNode(ProcessInstanceID, NextNode, PrevTaskNode)
		if err!=nil{
			return err
		}
	}

	return nil

}

//获取流程实例中某个Node 返回 Node
func GetInstanceNode(ProcessInstanceID int, NodeID string) (Node, error) {
	ProcID, err := GetProcessIDByInstanceID(ProcessInstanceID)
	if err != nil {
		return Node{}, err
	}

	//从Cache中获得流程节点列表
	Nodes, err := GetProcCache(ProcID)
	if err != nil {
		return Node{}, err
	}
	//获得节点
	node, ok := Nodes[NodeID]
	if !ok {
		return Node{}, fmt.Errorf("ID为%d的流程实例中不存在ID为%s的节点", ProcessInstanceID, NodeID)
	}

	return node, nil
}

//判断特定实例中某一个节点是否已经完成
func IfInstanceNodeIsFinish(ProcessInstanceID int, NodeID string) (bool, error) {
	var finished bool
	sql := "SELECT CASE WHEN total=finished THEN 1 ELSE 0 END AS finished " +
		"FROM " +
		"(SELECT COUNT(*) AS total,SUM(is_finished) AS finished " +
		"FROM `task` WHERE proc_inst_id=? AND node_id=? GROUP BY proc_inst_id,node_id) a"

	if _, err := dao.ExecSQL(sql, &finished, ProcessInstanceID, NodeID); err == nil {
		return finished, nil
	} else {
		return finished, err
	}
}
