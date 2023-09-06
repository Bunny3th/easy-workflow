package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/Bunny3th/easy-workflow/workflow/dao"
	. "github.com/Bunny3th/easy-workflow/workflow/model"
)

//生成任务 返回生成的任务ID数组
//思考，一个节点可能分配了N位用户，所以生成节点对应的Task的时候，也需要生成N条Task
//一个节点的上级节点可能不是一个，节点驳回的时候，就需要知道往哪个节点驳回,所以需要记录上一个节点是谁
func CreateTask(ProcessInstanceID int, NodeID string, PrevNodeID string, UserIDs []string) ([]int, error) {
	type user struct {
		UserID string
	}

	var users []user

	for _, u := range UserIDs {
		users = append(users, user{u})
	}

	j, err := json.Marshal(users)
	if err != nil {
		return nil, err
	}

	var r []int

	_, err = ExecSQL("call sp_task_create(?,?,?,?)", &r, ProcessInstanceID, NodeID, PrevNodeID, j)
	if err != nil {
		return nil, err
	}
	return r, nil
}

//task做通过、处理时可能会有一些附加功能，放在这里。
//目前只实现DirectlyToWhoRejectedMe，即任务通过时直接返回到上一个驳回我的节点。
//考虑这种情况：假设A、B、C、D、E 共5个任务节点节点，E节点（老板）使用自由驳回功能，直接驳回到A（员工），嗯就是这么任性
//传统情况下，A根据领导指示做修改重新提交后，B、C、D几个主管都要再审核一遍，来来回回，不仅效率低，由于增加工作量，各各一肚子怨气
//此时老板发话：芝麻绿豆大的事，B、C、D不用再参合了，小A你直接提我这边吧
//此时使用DirectlyToWhoRejectedMe，A直接提交到上次驳回他的E，皆大欢喜
//***需要注意的事，此功能只有在A是非会签节点时才能使用，否则试想，若节点中有甲乙两人，一人使用普通的pass，一人使用此时使用DirectlyToWhoRejectedMe，
//此时出现分歧，难道要打一架解决？
type taskOption struct {
	DirectlyToWhoRejectedMe bool //任务通过(pass)时直接返回到上一个驳回我的节点
}

//完成任务，在本节点处理完毕的情况下会自动处理下一个节点
func TaskPass(TaskID int, Comment string, VariableJson string, DirectlyToWhoRejectedMe bool) error {
	taskOption := taskOption{DirectlyToWhoRejectedMe: DirectlyToWhoRejectedMe}
	return taskHandle(TaskID, Comment, VariableJson, true, taskOption)
}

//驳回任务，在本节点处理完毕的情况下会自动处理下一个节点
func TaskReject(TaskID int, Comment string, VariableJson string) error {
	return taskHandle(TaskID, Comment, VariableJson, false, taskOption{})
}

//任务处理
func taskHandle(TaskID int, Comment string, VariableJson string, Pass bool, option taskOption) error {
	//获取节点信息
	task, err := GetTaskInfo(TaskID)
	if err != nil {
		return err
	}
	//判断节点是否已处理
	if task.IsFinished == 1 {
		return fmt.Errorf("节点ID%d已处理，无需操作", TaskID)
	}

	//判断是通过还是驳回
	var sql string
	if Pass == true { //通过
		sql = "call sp_task_pass(?,?,?,?)"
	} else { //驳回
		sql = "call sp_task_reject(?,?,?)"
		//获取task所在的node
		taskNode, err := GetInstanceNode(task.ProcInstID, task.NodeID)
		if err != nil {
			return err
		}
		//起始节点不能做驳回
		if taskNode.NodeType == RootNode {
			return errors.New("起始节点无法驳回!")
		}
	}

	type result struct {
		Error            string
		Next_opt_node_id string
	}
	var r result
	if Pass == true {
		_, err = ExecSQL(sql, &r, TaskID, Comment, VariableJson, option.DirectlyToWhoRejectedMe)
	} else {
		_, err = ExecSQL(sql, &r, TaskID, Comment, VariableJson)
	}

	if err != nil {
		return err
	}

	if r.Error != "" {
		return errors.New(r.Error)
	}
	//如果没有下一个节点要处理，直接退出
	if r.Next_opt_node_id == "" {
		return nil
	}

	//如果任务所在节点已处理完毕，此时：
	//1、处理节点结束事件
	//2、开始处理下一个节点
	task, err = GetTaskInfo(TaskID)
	if err != nil {
		return err
	}

	//需要处理的下一个节点
	NextNode, err := GetInstanceNode(task.ProcInstID, r.Next_opt_node_id)
	if err != nil {
		return err
	}

	//当前task所在节点
	CurrentNode, err := GetInstanceNode(task.ProcInstID, task.NodeID)
	if err != nil {
		return err
	}

	//当前task上一个节点.这里要注意，如果当前节点是开始节点，则上一个节点是空节点
	var PrevNode Node
	if CurrentNode.NodeType == RootNode {
		PrevNode = Node{}
	} else {
		PrevNode, err = GetInstanceNode(task.ProcInstID, task.PrevNodeID)
		if err != nil {
			return err
		}
	}

	//这里处理节点结束事件
	//由于存在会签节点，节点中可能有n个任务，如果每个任务处理后都要调用结束事件，显然有点不妥。结束事件应该只被执行一次
	//所以，这里先判断目前节点是否已经处理完毕。处理完毕，就可以开始节点结束事件。
	finished, err := InstanceNodeIsFinish(task.ProcInstID, CurrentNode.NodeID)
	if err != nil {
		return err
	}
	if finished {
		err = RunEvents(CurrentNode.EndEvents, task.ProcInstID, &CurrentNode, PrevNode)
		if err != nil {
			return err
		}
	}

	//开始处理下一个节点
	err = ProcessNode(task.ProcInstID, &NextNode, CurrentNode)
	if err != nil {
		return err
	}

	return nil
}

//获取任务信息
func GetTaskInfo(TaskID int) (Task, error) {
	var task Task
	_, err := ExecSQL("select * from task where id=?", &task, TaskID)
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

//获取特定用户待办任务列表
func GetTaskToDoList(UserID string) ([]Task, error) {
	var tasks []Task
	sql := "SELECT DISTINCT a.id,c.business_id,a.proc_id,d.name,a.proc_inst_id,a.node_id,b.node_name," +
		"a.user_id,a.create_time FROM task a " +
		"LEFT JOIN (SELECT DISTINCT proc_id,node_id,node_name FROM proc_execution) b " +
		"ON a.proc_id=b.proc_id AND a.node_id=b.node_id " +
		"LEFT JOIN proc_inst c ON a.proc_inst_id=c.id " +
		"LEFT JOIN proc_def d ON a.proc_id=d.id " +
		"WHERE a.user_id=?  AND a.is_finished=0 ORDER BY a.id;"

	_, err := ExecSQL(sql, &tasks, UserID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

//获取特定用户已完成任务列表
func GetTaskFinishedList(UserID string) ([]Task, error) {
	var tasks []Task
	sql := "WITH tmp_task AS " +
		"(SELECT id,proc_id,proc_inst_id,node_id,user_id,create_time,is_passed,is_finished,finished_time \n" +
		"FROM task \n" +
		"UNION ALL \n" +
		"SELECT task_id AS id,proc_id,proc_inst_id,node_id,user_id,create_time,is_passed,is_finished,finished_time \n" +
		"FROM hist_task)\n" +
		",tmp_task_comment AS \n" +
		"(SELECT task_id,COMMENT FROM task_comment \n" +
		"UNION ALL \n" +
		"SELECT task_id,COMMENT FROM hist_task_comment)\n" +
		",tmp_proc_inst AS \n" +
		"(SELECT id,business_id FROM proc_inst \n" +
		"UNION ALL \n" +
		"SELECT id,business_id FROM hist_proc_inst)\n" +
		",tmp_proc_execution AS \n" +
		"(SELECT DISTINCT proc_id,node_id,node_name FROM proc_execution) \n" +
		"SELECT DISTINCT a.id,c.business_id,a.proc_id,d.name,a.proc_inst_id,b.node_id,b.node_name,a.user_id,\n" +
		"DATE_FORMAT(a.create_time,'%Y-%m-%d %T') AS create_time,DATE_FORMAT(a.finished_time,'%Y-%m-%d %T') AS finished_time,\n" +
		"e.comment \n" +
		"FROM tmp_task a \n" +
		"LEFT JOIN tmp_proc_execution b ON a.proc_id=b.proc_id AND a.node_id=b.node_id \n" +
		"LEFT JOIN tmp_proc_inst c ON a.proc_inst_id=c.id \n" +
		"LEFT JOIN proc_def d ON a.proc_id=d.id \n" +
		"LEFT JOIN tmp_task_comment e ON a.id=e.task_id WHERE \n" +
		"a.user_id=?   AND a.is_finished=1  \n" +
		"AND a.is_passed IS NOT NULL \n" + //有些任务并不是用户自己完成，而是系统自动结束的，这种任务不必给用户看
		"ORDER BY a.id;"

	_, err := ExecSQL(sql, &tasks, UserID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

//根据流程定义,列出task所在节点的所有上流节点
func TaskUpstreamNodeList(TaskID int) ([]Node, error) {
	task, err := GetTaskInfo(TaskID)
	if err != nil {
		return nil, err
	}

	sql := "WITH RECURSIVE tmp(`node_id`,node_name,`prev_node_id`,`node_type`) AS " +
		"(SELECT `node_id`,node_name,`prev_node_id`,`node_type` " +
		"FROM `proc_execution` WHERE node_id=? " +
		"UNION ALL " +
		"SELECT a.`node_id`,a.node_name,a.`prev_node_id`,a.`node_type` " +
		"FROM `proc_execution` a JOIN tmp b ON a.node_id=b.`prev_node_id`) " +
		"SELECT node_id,node_name,prev_node_id,node_type FROM tmp WHERE node_type!=2 AND node_id!=?;"
	var nodes []Node
	if _, err := ExecSQL(sql, &nodes, task.NodeID, task.NodeID); err == nil {
		return nodes, nil
	} else {
		return nil, err
	}
}

//自由驳回到任意一个上游节点
func TaskFreeRejectToUpstreamNode(TaskID int, NodeID string, Comment string, VariableJson string) error {
	//思考:不论是否会签节点，自由驳回功能都可以使用。因为会签节点任意一人驳回就算驳回，其他人已经没有机会再做操作。

	type result struct {
		Error            string
		Next_opt_node_id string
	}
	var r result
	_, err := ExecSQL("call sp_task_reject(?,?,?)", &r, TaskID, Comment, VariableJson)
	if err != nil {
		return err
	}

	if r.Error != "" {
		return errors.New(r.Error)
	}

	task, err := GetTaskInfo(TaskID)
	if err != nil {
		return err
	}

	//当前task所在节点
	CurrentNode, err := GetInstanceNode(task.ProcInstID, task.NodeID)
	if err != nil {
		return err
	}

	//reject to 节点
	RejectToNode, err := GetInstanceNode(task.ProcInstID, NodeID)
	if err != nil {
		return err
	}

	err = ProcessNode(task.ProcInstID, &RejectToNode, CurrentNode)
	if err != nil {
		return err
	}
	return nil
}

//获取流程实例下任务历史记录
func GetInstanceTaskHistory(ProcessInstanceID int) ([]Task, error) {
	var tasklist []Task
	_, err := ExecSQL("call sp_proc_inst_task_hist(?)", &tasklist, ProcessInstanceID)
	if err != nil {
		return nil, err
	}
	return tasklist, nil
}
