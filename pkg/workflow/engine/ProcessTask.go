package engine

import (
	. "easy-workflow/pkg/dao"
	"encoding/json"
	"errors"

)

//生成任务 返回任务ID
//思考，一个节点可能分配了N位用户，所以生成节点对应的Task的时候，也需要生成N条Task
//一个节点的上级节点可能不是一个，节点驳回的时候，就需要知道往哪个节点驳回,所以需要记录上一个节点是谁
func CreateTask(ProcessInstanceID int, NodeID string,PrevNodeID string,UserIDs []string) ([]int, error) {

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

	_, err = ExecSQL("call sp_task_create(?,?,?,?)", &r, ProcessInstanceID, NodeID,PrevNodeID, j)
	if err != nil {
		return nil, err
	}
	return r, nil
}

//完成任务，并在本节点处理完毕的情况下返回下一个待处理节点(如果有)
func TaskPass(TaskID int, Comment string, VariableJson string) (string,error) {
	type result struct {
		Error string
		Next_opt_node_id string
	}
	var r result

	_,err:=ExecSQL("call sp_task_pass(?,?,?)",&r,TaskID,Comment,VariableJson)
	if err!=nil{
		return "",err
	}

	if r.Error!=""{
		return "",errors.New(r.Error)
	}




	return r.Next_opt_node_id,nil

}
