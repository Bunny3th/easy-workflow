package engine

import (
	"errors"
	"fmt"
	. "github.com/Bunny3th/easy-workflow/workflow/model"
	"regexp"
	"strings"
)

//处理节点,如：生成task、进行条件判断、处理结束节点等
func ProcessNode(ProcessInstanceID int, CurrentNode *Node, PrevNode Node) error {
	//这里处理开始事件
	err := RunNodeEvents(CurrentNode.NodeStartEvents, ProcessInstanceID, CurrentNode, PrevNode)
	if err != nil {
		return err
	}

	//开始节点也需要处理，因为开始节点可能因为驳回而重新回到开始节点，此时的开始节点=普通任务节点
	if CurrentNode.NodeType == RootNode {
		_, err := TaskNodeHandle(ProcessInstanceID, CurrentNode, PrevNode)
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
		_, err := TaskNodeHandle(ProcessInstanceID, CurrentNode, PrevNode)
		if err != nil {
			return err
		}
	}

	if CurrentNode.NodeType == EndNode {
		err := EndNodeHandle(ProcessInstanceID, 1)
		if err != nil {
			return err
		}
	}

	return nil
}

//开始节点处理 开始节点是一个特殊的任务节点，其特殊点在于:
//1、在生成流程实例的同时，就要运行开始节点
//2、开始节点生成的任务自动完成，而后自动进行下一个节点的处理
func startNodeHandle(ProcessInstanceID int, StartNode *Node, Comment string, VariableJson string) error {
	if StartNode.NodeType != RootNode {
		return errors.New("不是开始节点，无法处理节点:" + StartNode.NodeName)
	}

	//这里处理节点开始事件
	err := RunNodeEvents(StartNode.NodeStartEvents, ProcessInstanceID, StartNode, Node{})
	if err != nil {
		return err
	}

	//生成Task
	taskids, err := TaskNodeHandle(ProcessInstanceID, StartNode, Node{})
	if err != nil {
		return err
	}

	//完成task,并获取下一步NodeID
	err = TaskPass(taskids[0], Comment, VariableJson, false)
	if err != nil {
		return err
	}

	return nil
}

//结束节点处理 结束节点只做收尾工作，将数据库中此流程实例产生的数据归档
//Status 流程实例状态 1:已完成 2:撤销
func EndNodeHandle(ProcessInstanceID int, Status int) error {
	//开启事务
	tx := DB.Begin()

	//***这里注意，经过多次测试，执行原生SQL，无返回值必须用Exec,用Raw会不执行***

	//将task表中所有该流程未finish的设置为finish
	result := tx.Exec("UPDATE proc_task SET is_finished=1,finished_time=NOW() "+
		"WHERE proc_inst_id=? AND is_finished=0;", ProcessInstanceID)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	//将task表中任务归档
	result = tx.Exec("INSERT INTO hist_proc_task(task_id,proc_id,proc_inst_id,\n"+
		"business_id,starter,node_id,node_name,prev_node_id,is_cosigned,\n"+
		"batch_code,user_id,`status`,is_finished,`comment`,proc_inst_create_time,create_time,finished_time)\n "+
		"SELECT id,proc_id,proc_inst_id,business_id,starter,\n"+
		"node_id,node_name,prev_node_id,is_cosigned,batch_code,user_id,`status`,\n"+
		"is_finished,`comment`,proc_inst_create_time,create_time,finished_time \n"+
		"FROM proc_task WHERE proc_inst_id=?;", ProcessInstanceID)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	//删除task表中历史数据
	result = tx.Exec(" DELETE FROM proc_task WHERE proc_inst_id=?;", ProcessInstanceID)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	//更新proc_inst表中状态
	result = tx.Exec("UPDATE proc_inst SET `status`=? WHERE id=?;", Status, ProcessInstanceID)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	//将proc_inst表中数据归档
	result = tx.Exec("INSERT INTO hist_proc_inst(proc_inst_id,proc_id,proc_version,business_id,starter,current_node_id,create_time,`status`)\n        "+
		"SELECT id,proc_id,proc_version,business_id,starter,current_node_id,create_time,`status`\n        "+
		"FROM proc_inst \n        "+
		"WHERE id=?; ", ProcessInstanceID)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	//删除proc_inst表中历史数据
	result = tx.Exec("DELETE FROM proc_inst WHERE id=?;", ProcessInstanceID)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	//将proc_inst_variable表中数据归档
	result = tx.Exec("INSERT INTO hist_proc_inst_variable(proc_inst_id,`key`,`value`)\n"+
		"SELECT proc_inst_id,`key`,`value` FROM proc_inst_variable WHERE proc_inst_id=?;", ProcessInstanceID)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	//删除proc_inst_variable表中历史数据
	result = tx.Exec("DELETE FROM proc_inst_variable WHERE proc_inst_id=?;", ProcessInstanceID)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}
	//提交事务
	tx.Commit()

	return nil
}

//任务节点处理 返回生成的taskid数组
func TaskNodeHandle(ProcessInstanceID int, CurrentNode *Node, PrevNode Node) ([]int, error) {
	//获取节点用户
	users, err := resolveNodeUser(ProcessInstanceID, *CurrentNode)
	if err != nil {
		return nil, err
	}

	//如果没有处理人，则任务无法分配
	if len(users) < 1 {
		return nil, errors.New("未指定处理人，无法处理节点:" + CurrentNode.NodeName)
	}

	//开始节点只能有一个用户发起,不管多少用户，只要第一个
	//思考：如果开始节点有多个处理人，则可能进入会签状态，可能造成流程都无法正常开始
	//所以，开始节点只能有一个处理人，且默认就是非会签节点
	if CurrentNode.NodeType == RootNode {
		users = users[0:1]
	}

	//生成Task
	taskIDs, err := CreateTask(ProcessInstanceID, CurrentNode.NodeID, PrevNode.NodeID, users)
	if err != nil {
		return nil, err
	}
	return taskIDs, nil
}

//GateWay节点处理
func GateWayNodeHandle(ProcessInstanceID int, CurrentNode *Node, PrevTaskNode Node) error {
	//--------------------首先，混合节点需要确认所有的上级节点都处理完，才能做下一步--------------------
	var totalFinished int                          //所有已完成的上级节点
	totalPrevNodes := len(CurrentNode.PrevNodeIDs) //所有上级节点
	for _, nodeID := range CurrentNode.PrevNodeIDs {
		finished, err := InstanceNodeIsFinish(ProcessInstanceID, nodeID)
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
		ok, err := ExpressionEvaluator(expression)
		if err != nil {
			return err
		}
		if ok {
			conditionNodeIDs = append(conditionNodeIDs, c.NodeID)
		}
	}

	//-------将conditionNodeIDs和InevitableNodes中的值一起放入nextNodeIDs，这是真正要处理的节点ID-------
	//去重(节点ID如果重复，意味着一个节点要做N次处理，这是灾难)
	nextNodeIDs := MakeUnique(conditionNodeIDs, CurrentNode.GWConfig.InevitableNodes)

	//这里处理节点结束事件
	err := RunNodeEvents(CurrentNode.NodeEndEvents, ProcessInstanceID, CurrentNode, PrevTaskNode)
	if err != nil {
		return err
	}

	//------------------------------对下级节点进行处理------------------------------
	for _, nodeID := range nextNodeIDs {
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
		err = ProcessNode(ProcessInstanceID, &NextNode, PrevTaskNode)
		if err != nil {
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
//注意，finish只是代表节点是不是已经处理，不管处理的方式是驳回还是通过
//一个流程实例中，由于驳回等原因，x节点可能出现多次。这里使用统计所有x节点的任务是否都finish来判断x节点是否finish
func InstanceNodeIsFinish(ProcessInstanceID int, NodeID string) (bool, error) {
	var finished bool
	sql := "SELECT CASE WHEN total=finished THEN 1 ELSE 0 END AS finished " +
		"FROM " +
		"(SELECT COUNT(*) AS total,SUM(is_finished) AS finished " +
		"FROM `proc_task` WHERE proc_inst_id=? AND node_id=? GROUP BY proc_inst_id,node_id) a"

	if _, err := ExecSQL(sql, &finished, ProcessInstanceID, NodeID); err == nil {
		return finished, nil
	} else {
		return finished, err
	}
}

//解析节点用户
//1、获得用户变量
//2、用户去重
func resolveNodeUser(ProcessInstanceID int, node Node) ([]string, error) {
	//匹配节点用户变量
	kv, err := ResolveVariables(ProcessInstanceID, node.UserIDs)
	if err != nil {
		return nil, err
	}

	//使用map去重，因为有可能某几个变量指向同一个用户，重复的用户会产生重复的任务
	var usersMap = make(map[string]string)
	for _, v := range kv {
		usersMap[v] = ""
	}

	//生成user数组
	var users []string
	for k, _ := range usersMap {
		users = append(users, k)
	}

	return users, nil
}
