package process

import (
	"easy-workflow/pkg/dao"
	. "easy-workflow/pkg/workflow/model/node"
	"easy-workflow/pkg/workflow/util"
	"errors"
)

//流程定义解析
func ProcessParse(Resource string) ([]Node, error) {
	var nodes []Node

	err := util.Json2Struct(Resource, &nodes)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

//流程定义保存
func ProcessSave(ProcessName string, Resource string, CreateUserID string, Source string) error {
	if ProcessName == "" || Source == "" || CreateUserID == "" {
		return errors.New("流程名称、来源、创建人ID不能为空")
	}

	_, err := ProcessParse(Resource)
	if err != nil {
		return err
	}

	type result struct {
		Error string
	}

	var r result
	_, err = dao.ExecSQL("CALL sp_proc_def_save(?,?,?,?)", &r, ProcessName, Resource, CreateUserID, Source)

	if err != nil || r.Error != "" {
		return errors.New(err.Error() + r.Error)
	}
	return nil
}

//获取流程ID
func GetProcessID(ProcessName string, Source string) (int, error) {
	var ID int
	_, err := dao.ExecSQL("SELECT id FROM proc_def where name=? and source=?", &ID, ProcessName, Source)

	if err != nil {
		return 0, err
	}

	return ID, nil
}

//获取流程的所有节点
func GetProcessNodes(ProcessID int) ([]Node, error) {
	return nil, nil
}
