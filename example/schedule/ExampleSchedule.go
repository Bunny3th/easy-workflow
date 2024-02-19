package schedule

import (
	. "github.com/Bunny3th/easy-workflow/workflow/engine"
	. "github.com/Bunny3th/easy-workflow/workflow/model"
)

//这里定义一个任务计划:对于UserID为"-1"的任务做自动通过
func AutoFinishTask() error {
	//首先获取所有用户ID为"-1"，且还未完成的任务
	var tasks []Task
	sql := "SELECT * FROM proc_task WHERE user_id='-1' AND is_finished=0"
	_, err := ExecSQL(sql, &tasks)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		//所谓自动通过，可以这么理解:1、上一个节点是通过状态,那我也就随大流通过;2、上一个节点做了驳回，我也随大流驳回
		//所以，需要根据本任务上一个节点，判断是应该通过还是驳回
		//首先得到流程定义中本节点的上一级节点数组
		type node struct {
			NodeID string `gorm:"column:node_id"`
		}
		var PrevNodes []node
		var PrevNodeIDs = make(map[string]any)
		_, err := ExecSQL("SELECT prev_node_id AS node_id FROM proc_execution WHERE proc_id=? AND node_id=?",
			&PrevNodes,
			task.ProcID, task.NodeID)
		if err != nil {
			return err
		}
		//节点ID加入Map中
		for _, n := range PrevNodes {
			PrevNodeIDs[n.NodeID] = nil
		}

		if _, ok := PrevNodeIDs[task.PrevNodeID]; ok {
			//如果本任务的上一个节点是流程定义中上一个节点，说明上一个节点是做了通过的,否则不可能到我这里,则我也通过
			err :=TaskPass(task.TaskID, "免审自动通过", "", false)
			if err != nil {
				return err
			}
		} else { //否则驳回
			err := TaskReject(task.TaskID, "免审自动驳回", "")
			if err != nil {
				return err
			}
		}

	}
	return nil
}
