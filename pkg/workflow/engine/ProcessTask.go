package engine

import (
	. "easy-workflow/pkg/dao"
	. "easy-workflow/pkg/workflow/model/datatables"
	"encoding/json"
	"errors"
	"fmt"
)

//生成任务 返回任务ID
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

	//type ID int
	var r []int

	_, err = ExecSQL("call sp_task_create(?,?,?,?)", &r, ProcessInstanceID, NodeID, PrevNodeID, j)
	if err != nil {
		return nil, err
	}
	return r, nil
}

//完成任务，并在本节点处理完毕的情况下返回下一个待处理节点(如果有)
func TaskPass(TaskID int, Comment string, VariableJson string) (error) {
	//type result struct {
	//	Error            string
	//	Next_opt_node_id string
	//}
	//var r result
	//
	//_, err := ExecSQL("call sp_task_pass(?,?,?)", &r, TaskID, Comment, VariableJson)
	//if err != nil {
	//	return err
	//}
	//
	//if r.Error != "" {
	//	return errors.New(r.Error)
	//}
	//
	//task,err:=GetTaskInfo(TaskID)
	//if err != nil {
	//	return err
	//}
	//
	////
	//NextNode,ok,err:=GetInstanceNode(task.ProcInstID,r.Next_opt_node_id)
	//if err != nil {
	//	return err
	//}
	//
	////当前节点
	//CurrentNode,_,err:=GetInstanceNode(task.ProcInstID,task.NodeID)
	//if err != nil {
	//	return err
	//}
	//
	//if ok{
	//	err=ProcessNode(task.ProcInstID,NextNode,CurrentNode)
	//	if err != nil {
	//		return err
	//	}
	//
	//}else{
	//	return fmt.Errorf ("无法找到节点ID:%d 的下一个节点",TaskID)
	//}
	//
	//return nil
	return taskHandle(TaskID, Comment, VariableJson,true)
}


//完成任务，并在本节点处理完毕的情况下返回下一个待处理节点(如果有)
func TaskReject(TaskID int, Comment string, VariableJson string) (error) {
	return taskHandle(TaskID, Comment, VariableJson,false)
}


func taskHandle(TaskID int, Comment string, VariableJson string,Pass bool) error{
	type result struct {
		Error            string
		Next_opt_node_id string
	}
	var r result

	var sql string
	if Pass==true{
		sql="call sp_task_pass(?,?,?)"
	}else{
		sql="call sp_task_reject(?,?,?)"
	}

	_, err := ExecSQL(sql, &r, TaskID, Comment, VariableJson)
	if err != nil {
		return err
	}

	if r.Error != "" {
		return errors.New(r.Error)
	}

	task,err:=GetTaskInfo(TaskID)
	if err != nil {
		return err
	}

	//
	NextNode,ok,err:=GetInstanceNode(task.ProcInstID,r.Next_opt_node_id)
	if err != nil {
		return err
	}

	//当前节点
	CurrentNode,_,err:=GetInstanceNode(task.ProcInstID,task.NodeID)
	if err != nil {
		return err
	}

	if ok{
		err=ProcessNode(task.ProcInstID,NextNode,CurrentNode)
		if err != nil {
			return err
		}

	}else{
		return fmt.Errorf ("无法找到TaskID:%d 所在节点 %s 的下一个待处理节点",TaskID,task.NodeID)
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
