package engine

import (
	. "easy-workflow/pkg/dao"
	. "easy-workflow/pkg/workflow/model"
	"encoding/json"
	"errors"
	"fmt"
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

//完成任务，在本节点处理完毕的情况下会自动处理下一个节点
func TaskPass(TaskID int, Comment string, VariableJson string) error {
	return taskHandle(TaskID, Comment, VariableJson, true)
}

//驳回任务，在本节点处理完毕的情况下会自动处理下一个节点
func TaskReject(TaskID int, Comment string, VariableJson string) error {
	return taskHandle(TaskID, Comment, VariableJson, false)
}

//任务处理
func taskHandle(TaskID int, Comment string, VariableJson string, Pass bool) error {
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
	if Pass == true {
		sql = "call sp_task_pass(?,?,?)"
	} else {
		sql = "call sp_task_reject(?,?,?)"
	}

	type result struct {
		Error            string
		Next_opt_node_id string
	}
	var r result
	_, err = ExecSQL(sql, &r, TaskID, Comment, VariableJson)
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

	err = ProcessNode(task.ProcInstID, NextNode, CurrentNode)
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
	_, err := ExecSQL("call sp_task_todo(?)", &tasks, UserID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

//获取特定用户已完成任务列表
func GetTaskFinishedList(UserID string) ([]Task, error) {
	var tasks []Task
	_, err := ExecSQL("call sp_task_finished(?)", &tasks, UserID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}