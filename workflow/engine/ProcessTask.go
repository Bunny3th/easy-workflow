package engine

import (
	//"encoding/json"
	"errors"
	"fmt"
	. "github.com/Bunny3th/easy-workflow/workflow/dao"
	"github.com/Bunny3th/easy-workflow/workflow/database"
	. "github.com/Bunny3th/easy-workflow/workflow/model"
	"time"
)

//生成任务 返回生成的任务ID数组
//思考，一个节点可能分配了N位用户，所以生成节点对应的Task的时候，也需要生成N条Task
//一个节点的上级节点可能不是一个，节点驳回的时候，就需要知道往哪个节点驳回,所以需要记录上一个节点是谁
func CreateTask(ProcessInstanceID int, NodeID string, PrevNodeID string, UserIDs []string) ([]int, error) {

	/*考虑以下情况：
	      |--C
	A--B--|
	      |--D
	假如C驳回到B,B重新提交，D是不是要处理两遍任务？（第1遍是B初次提交后的任务，第2遍是B再次提交后重新生成的任务）
	所以，在同一实例中，如果某一个节点任务还未finish，这个节点任务就不应该再次被生成
	*/
	var task Task
	result := DB.Raw("SELECT * FROM task "+
		"WHERE proc_inst_id=? AND node_id=? AND is_finished=0 "+
		"ORDER BY `task`.`id` LIMIT 1", ProcessInstanceID, NodeID).Scan(&task)
	if result.Error != nil {
		return nil, result.Error
	}
	if task.TaskID != 0 {
		return nil, nil
	}

	//获取流程ID
	ProcID, err := GetProcessIDByInstanceID(ProcessInstanceID)
	if err != nil {
		return nil, err
	}

	//获取Node信息
	Node, err := GetInstanceNode(ProcessInstanceID, NodeID)
	if err != nil {
		return nil, err
	}

	//生成批次码
	var BatchCode string
	_, err = ExecSQL("SELECT UUID()", &BatchCode)
	if err != nil {
		return nil, err
	}

	//开始生成数据
	var tasks []database.Task
	for _, u := range UserIDs {
		tasks = append(tasks, database.Task{ProcID: ProcID, ProcInstID: ProcessInstanceID, NodeID: NodeID,
			PrevNodeID: PrevNodeID, IsCosigned: Node.IsCosigned, BatchCode: BatchCode, UserID: u})
	}

	//开启事务
	tx := DB.Begin()

	//Task存入数据库
	result = tx.Create(&tasks)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	//更新proc_inst表`current_node_id`字段
	result = tx.Model(&database.ProcInst{}).Where("id=?", ProcessInstanceID).Update("current_node_id", NodeID)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	//提交事务
	tx.Commit()

	//获取存入数据库后生成的TaskID
	var TaskIDs []int
	for _, t := range tasks {
		TaskIDs = append(TaskIDs, t.ID)
	}

	return TaskIDs, nil
}

//task处理时可能会有一些附加功能，放在这里。
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
	//获取节点信息
	taskInfo, err := GetTaskInfo(TaskID)
	if err != nil {
		return err
	}
	//判断节点是否已处理
	if taskInfo.IsFinished == 1 {
		return fmt.Errorf("节点ID%d已处理，无需操作", TaskID)
	}

	//------------------------	DirectlyToWhoRejectedMe 功能前置验证 ------------------------
	//1、是否是会签节点
	//2、是否存在上一个任务节点?上一个节点是否做的是驳回
	if DirectlyToWhoRejectedMe {
		//会签节点无法使用此功能，因为会签节点没有“统一意志”
		if taskInfo.IsCosigned == 1 {
			return errors.New("会签节点无法使用【DirectlyToWhoRejectedMe】功能!")
		}

		//判断任务的上一个节点是不是做了驳回
		err, PrevNodeIsReject := taskPrevNodeIsReject(taskInfo)
		if err != nil {
			return err
		}

		if PrevNodeIsReject == false {
			return errors.New("此任务的上一节点并未做驳回,无法使用【DirectlyToWhoRejectedMe】功能！")
		}
	}

	//任务提交数据保存
	err = taskSubmitSave(TaskID, Comment, VariableJson, 1)
	if err != nil {
		return err
	}

	//完成任务后的后继处理
	err = ProcessAfterTaskComplete(TaskID, taskOption{DirectlyToWhoRejectedMe: DirectlyToWhoRejectedMe})
	if err != nil {
		return err
	}

	return nil
}

//驳回任务，在本节点处理完毕的情况下会自动处理下一个节点
func TaskReject(TaskID int, Comment string, VariableJson string) error {
	//获取节点信息
	taskInfo, err := GetTaskInfo(TaskID)
	if err != nil {
		return err
	}
	//判断节点是否已处理
	if taskInfo.IsFinished == 1 {
		return fmt.Errorf("节点ID%d已处理，无需操作", TaskID)
	}

	//获取task所在的node
	taskNode, err := GetInstanceNode(taskInfo.ProcInstID, taskInfo.NodeID)
	if err != nil {
		return err
	}
	//起始节点不能做驳回
	if taskNode.NodeType == RootNode {
		return errors.New("起始节点无法驳回!")
	}

	//任务提交数据保存
	err = taskSubmitSave(TaskID, Comment, VariableJson, 2)
	if err != nil {
		return err
	}

	//完成任务后的后继处理
	err = ProcessAfterTaskComplete(TaskID, taskOption{})
	if err != nil {
		return err
	}

	return nil
}

//任务完成后的处理
func ProcessAfterTaskComplete(TaskID int, option taskOption) error {
	////获取节点信息
	taskInfo, err := GetTaskInfo(TaskID)
	if err != nil {
		return err
	}

	//获取任务执行完毕后下一个节点
	var NextNode Node
	//如果任务动作是“pass” and 开启 DirectlyToWhoRejectedMe,直接使用任务的PrevNodeID
	if taskInfo.Status == 1 && option.DirectlyToWhoRejectedMe {
		NextNode, err = GetInstanceNode(taskInfo.ProcInstID, taskInfo.PrevNodeID)
		if err != nil {
			return err
		}
	} else { //否则就通过计算得出下一个节点是谁
		NextNode, err = TaskNextNode(taskInfo.TaskID)
		if err != nil {
			return err
		}
	}

	//如果需要处理的下一个节点是一个空Node，则说明:当前任务节点还没有处理完,直接退出
	if NextNode.NodeID == "" {
		return nil
	}

	//执行到这一步,说明所在节点中所有任务已全部处理完毕，此时：
	//1、处理节点结束事件
	//2、开始处理下一个节点

	//当前task所在节点
	CurrentNode, err := GetInstanceNode(taskInfo.ProcInstID, taskInfo.NodeID)
	if err != nil {
		return err
	}

	//当前task上一个节点.这里要注意，如果当前节点是开始节点，则上一个节点是空节点
	var PrevNode Node
	if CurrentNode.NodeType == RootNode {
		PrevNode = Node{}
	} else {
		PrevNode, err = GetInstanceNode(taskInfo.ProcInstID, taskInfo.PrevNodeID)
		if err != nil {
			return err
		}
	}

	//--------------------------这里处理节点结束事件--------------------------
	err = RunEvents(CurrentNode.EndEvents, taskInfo.ProcInstID, &CurrentNode, PrevNode)
	if err != nil {
		return err
	}

	//--------------------------开始处理下一个节点--------------------------
	err = ProcessNode(taskInfo.ProcInstID, &NextNode, CurrentNode)
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
	if task.TaskID == 0 {
		return Task{}, fmt.Errorf("ID为%d的任务不存在!", TaskID)
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
		"(SELECT id,proc_id,proc_inst_id,node_id,user_id,create_time,`status`,is_finished,finished_time \n" +
		"FROM task \n" +
		"UNION ALL \n" +
		"SELECT task_id AS id,proc_id,proc_inst_id,node_id,user_id,create_time,`status`,is_finished,finished_time \n" +
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
		"AND a.`status`!=0 \n" + //有些任务并不是用户自己完成，而是系统自动结束的，此时任务状态还是0，这种任务不必给用户看
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

	taskInfo, err := GetTaskInfo(TaskID)
	if err != nil {
		return err
	}

	//判断节点是否已处理
	if taskInfo.IsFinished == 1 {
		return fmt.Errorf("节点ID%d已处理，无需操作", TaskID)
	}

	//获取task所在的node
	taskNode, err := GetInstanceNode(taskInfo.ProcInstID, taskInfo.NodeID)
	if err != nil {
		return err
	}
	//起始节点不能做驳回
	if taskNode.NodeType == RootNode {
		return errors.New("起始节点无法驳回!")
	}

	//保存数据
	taskSubmitSave(TaskID, Comment, VariableJson, 2)

	//当前task所在节点
	CurrentNode, err := GetInstanceNode(taskInfo.ProcInstID, taskInfo.NodeID)
	if err != nil {
		return err
	}

	//reject to 节点
	RejectToNode, err := GetInstanceNode(taskInfo.ProcInstID, NodeID)
	if err != nil {
		return err
	}

	err = ProcessNode(taskInfo.ProcInstID, &RejectToNode, CurrentNode)
	if err != nil {
		return err
	}
	return nil
}

//获取流程实例下任务历史记录
func GetInstanceTaskHistory(ProcessInstanceID int) ([]Task, error) {
	var tasklist []Task
	sql := "WITH tmp_task AS(\n" +
		"SELECT task_id AS id,proc_id,proc_inst_id,node_id,prev_node_id,\n        " +
		"is_cosigned,batch_code,user_id,`status`,is_finished,create_time,finished_time\n        " +
		"FROM hist_task\n        " +
		"WHERE proc_id=?\n        " +
		"UNION ALL\n        " +
		"SELECT id,proc_id,proc_inst_id,node_id,prev_node_id,\n        " +
		"is_cosigned,batch_code,user_id,`status`,is_finished,create_time,finished_time \n        " +
		"FROM task   \n        " +
		"WHERE proc_id=?\n    " +
		"),tmp_proc AS(\n        " +
		"SELECT id,business_id FROM proc_inst WHERE id=?\n        " +
		"UNION ALL\n        " +
		"SELECT proc_inst_id AS id,business_id FROM hist_proc_inst WHERE proc_inst_id=?\n    " +
		"),tmp_task_comment AS (\n        " +
		"SELECT task_id,COMMENT FROM task_comment\n        " +
		"UNION ALL\n        " +
		"SELECT task_id,COMMENT FROM hist_task_comment    \n    " +
		"),tmp_proc_inst AS (\n        " +
		"SELECT id,business_id FROM proc_inst\n        " +
		"UNION ALL\n        " +
		"SELECT id,business_id FROM hist_proc_inst    \n    " +
		")\n    \n    " +
		"SELECT DISTINCT\n    " +
		"a.id,\n    " +
		"a.proc_id,\n    " +
		"a.proc_inst_id,\n    " +
		"c.business_id,\n    " +
		"a.node_id,\n    " +
		"b.node_name,\n    " +
		"a.is_cosigned,\n    " +
		"a.batch_code,\n    " +
		"a.user_id,\n    " +
		"a.`status`,\n    " +
		"a.is_finished,\n    " +
		"d.comment,\n    " +
		"DATE_FORMAT(a.create_time,'%Y-%m-%d %T')  AS create_time,\n    " +
		"DATE_FORMAT(a.finished_time,'%Y-%m-%d %T') AS finished_time    \n    " +
		"FROM \n    " +
		"tmp_task a\n    " +
		"LEFT JOIN proc_execution b ON a.node_id=b.node_id\n    " +
		"LEFT JOIN tmp_proc_inst c ON a.proc_inst_id=c.id\n    " +
		"LEFT JOIN tmp_task_comment d ON a.id=d.task_id\n    " +
		"ORDER BY a.id;"
	_, err := ExecSQL(sql, &tasklist, ProcessInstanceID, ProcessInstanceID, ProcessInstanceID, ProcessInstanceID)
	if err != nil {
		return nil, err
	}
	return tasklist, nil
}

//获取任务执行完毕后下一个节点
func TaskNextNode(TaskID int) (Node, error) {
	taskInfo, err := GetTaskInfo(TaskID)
	if err != nil {
		return Node{}, err
	}

	//获取任务节点审批状态
	TotalTask, TotalPassed, _, err := TaskNodeStatus(TaskID)
	if err != nil {
		return Node{}, err
	}

	//----------------------------------首先判断任务通过的情况----------------------------------
	if taskInfo.Status == 1 {
		//1、非会签且通过
		//2、会签且通过,目前节点中通过数与总任务数一致
		//以上两种情况，直接流向“流程定义中该任务节点的下一个节点”
		if (taskInfo.IsCosigned == 0) ||
			(taskInfo.IsCosigned == 1 && TotalTask == TotalPassed) {
			//流程定义中该任务节点的下一个节点,注意:
			//任务节点的下一个节点永远只有一个，一个任务节点下不可能直接衍生出多个任务节点,因为衍生多个节点是网关的任务
			//而本项目中是混合网关，任务节点下没有必要生成多个网关
			var ProcExecutionNextNode database.ProcExecution
			result := DB.Where("prev_node_id=?", taskInfo.NodeID).First(&ProcExecutionNextNode)
			if result.Error != nil {
				return Node{}, result.Error
			}
			return GetInstanceNode(taskInfo.ProcInstID, ProcExecutionNextNode.NodeID)
		} else {
			return Node{}, nil
		}
	}

	//----------------------------------接下来判断在驳回的情况下，应该到哪个节点----------------------------------
	if taskInfo.Status == 2 {
		/*
			# 注意，由于有自由驳回与DirectlyReturnToWhoRejectedMe(简称DR)两个能跨节点的功能，这里需要做比较复杂的判断
			# 考虑这样一种情况 A、B、C、D、E 五个节点：
			# E直接驳回到A，A使用DR重新提交到E，此时E再驳回，肯定不是想驳回到流程定义中的D，而是驳回到A
			# 此时task表中E任务的prev_node_id 字段为A,正是E想要驳回到的节点
			# 那么驳回时判断下一节点就直接使用 prev_node_id 字段？不行！考虑以下情况:
			# E驳回到D，此时D任务的prev_node_id是E，D再驳回，难道驳回到E？
			# 所以，节点驳回后应到哪个节点，应该分两步判断：
			# 1、如果X节点是由Y节点通过后流程到达，则X驳回时直接用 prev_node_id 字段
			# 2、如果X节点是由Y节点驳回后流程到达，则X驳回应使用流程定义中的"上一个节点"
		*/
		var PrevNodeID string

		//上一个节点是否做了驳回
		err, PrevNodeIsReject := taskPrevNodeIsReject(taskInfo)
		if err != nil {
			return Node{}, err
		}

		//上一个节点中没有做驳回
		if PrevNodeIsReject == false {
			//直接使用本任务的prev_node_id字段作为上一个节点
			PrevNodeID = taskInfo.PrevNodeID
		} else {
			//若上一个节点做了驳回，则用递归逆推获得流程定义中上一个节点
			Nodes, err := TaskUpstreamNodeList(TaskID)
			if err != nil {
				return Node{}, err
			}

			PrevNodeID = Nodes[0].NodeID
		}

		//不管是否会签，驳回都会返回上一个节点
		return GetInstanceNode(taskInfo.ProcInstID, PrevNodeID)
	}

	return Node{}, nil
}

//任务节点审批状态 返回节点总任务数量、通过数、驳回数、
func TaskNodeStatus(TaskID int) (int, int, int, error) {
	taskInfo, err := GetTaskInfo(TaskID)
	if err != nil {
		return 0, 0, 0, err
	}

	type Result struct {
		TotalTask     int `gorm:"column:total_task"`
		TotalPassed   int `gorm:"column:total_passed"`
		TotalRejected int `gorm:"column:total_rejected"`
	}
	var result Result

	r := DB.Raw("SELECT COUNT(*) AS total_task,\n"+
		"SUM(CASE `status` WHEN 1 THEN 1 ELSE 0 END) AS total_passed,\n"+
		"SUM(CASE `status` WHEN 2 THEN 1 ELSE 0 END) AS total_rejected \n"+
		"FROM task \n"+
		"WHERE proc_inst_id=? \n "+
		"AND node_id=? \n  "+
		"AND `batch_code`=?;", taskInfo.ProcInstID, taskInfo.NodeID, taskInfo.BatchCode).Scan(&result)
	if r.Error != nil {
		return 0, 0, 0, r.Error
	}

	return result.TotalTask, result.TotalPassed, result.TotalRejected, nil
}

//将任务提交数据(通过、驳回、变量)保存到数据库
func taskSubmitSave(TaskID int, Comment string, VariableJson string, Status int) error {
	taskInfo, err := GetTaskInfo(TaskID)
	if err != nil {
		return err
	}
	//判断节点是否已处理
	if taskInfo.IsFinished == 1 {
		return fmt.Errorf("节点ID%d已处理，无需操作", TaskID)
	}

	//设置实例变量
	err = InstanceVariablesSave(taskInfo.ProcInstID, VariableJson)
	if err != nil {
		return err
	}

	//------------------------------------开始事务------------------------------------
	tx := DB.Begin()

	//更新task表记录
	result := tx.Model(&database.Task{}).Where("id=?", taskInfo.TaskID).
		Updates(database.Task{Status: Status, IsFinished: 1, FinishedTime: time.Now()})
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	//1、非会签节点，一人通过即通过，所以要把其他人的任务finish掉
	//2、不论是否会签，都是一人驳回即驳回，所以需要把同一批次task的isfinish设置为1,让其他人不用再处理
	if (taskInfo.IsCosigned == 0 && Status == 1) || Status == 2 {
		result = tx.Model(&database.Task{}).Where("batch_code=?", taskInfo.BatchCode).
			Updates(database.Task{IsFinished: 1, FinishedTime: time.Now()})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}

	//生成comment记录
	if Comment != "" {
		result = tx.Create(&database.TaskComment{TaskID: taskInfo.TaskID, ProcInstID: taskInfo.ProcInstID, Comment: Comment})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	//事务提交
	tx.Commit()

	return nil
}

//任务的上一个节点是不是做了驳回
func taskPrevNodeIsReject(TaskInfo Task) (error, bool) {
	//获得实际执行过程中上一个节点的BatchCode
	type BatchCode struct {
		BatchCode string `gorm:"column:batch_code"`
	}
	var batchCode BatchCode

	DB.Raw("SELECT a.batch_code\n"+
		"FROM task a\n "+
		"JOIN \n"+
		"(SELECT prev_node_id,proc_inst_id FROM task WHERE id=?) b \n        "+
		"ON a.node_id=b.prev_node_id AND a.proc_inst_id=b.proc_inst_id\n        "+
		"ORDER BY a.id DESC LIMIT 1;", TaskInfo.TaskID).Scan(&batchCode)

	if batchCode.BatchCode == "" {
		return errors.New("此任务不存在上一节点!"), false
	}

	//获得上一个实际执行过程中状态为驳回的任务
	var prevTask database.Task
	result := DB.Raw("SELECT id FROM task WHERE batch_code =? AND `status`=2 LIMIT 1", batchCode.BatchCode).Scan(&prevTask)
	if result.Error != nil {
		return result.Error, false
	}

	//没有找到,说明上一个节点中没有做驳回
	if prevTask.ID == 0 {
		return nil, false
	} else {
		return nil, true
	}
}

//此方法方便前端判断，某一个任务可以执行哪些操作
//目前为止，除了传统的通过驳回，本项目还增加了"自由驳回"与"直接提交到上一个驳回我的节点"
//而"直接提交到上一个驳回我的节点"：
//1、在会签节点无法使用 2、在此任务的上一节点并未做驳回时也无法使用
//对于前端而言，实现无法提前知道这些信息。
//难道让用户一个一个点按钮试错？此方法目的是解决这个困扰
func WhatCanIDo(TaskID int) (TaskAction, error) {
	var act TaskAction
	act = TaskAction{CanPass: true, CanReject: true, CanFreeRejectToUpstreamNode: true, CanDirectlyToWhoRejectedMe: true} //初始化

	taskInfo, err := GetTaskInfo(TaskID)
	if err != nil {
		return TaskAction{}, err
	}

	node, err := GetInstanceNode(taskInfo.ProcInstID, taskInfo.NodeID)
	if err != nil {
		return TaskAction{}, nil
	}

	//起始节点不能做驳回动作
	if node.NodeType == RootNode {
		act.CanReject = false
		act.CanFreeRejectToUpstreamNode = false
	}

	//会签节点不能使用DirectlyToWhoRejectedMe功能
	if taskInfo.IsCosigned == 1 {
		act.CanDirectlyToWhoRejectedMe = false
	}

	//此任务的上一节点并未做驳回,无法使用DirectlyToWhoRejectedMe功能
	err, PrevNodeIsReject := taskPrevNodeIsReject(taskInfo)
	if err != nil {
		return TaskAction{}, err
	}
	if PrevNodeIsReject == false {
		act.CanDirectlyToWhoRejectedMe = false
	}

	return act, nil
}
