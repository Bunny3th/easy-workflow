package model

type ProcessDefine struct {
	ID         int    `gorm:"column:id"`          //流程ID
	Name       string `gorm:"column:name"`        //流程名字
	Version    int    `gorm:"column:version"`     //版本号
	Resource   string `gorm:"column:resource"`    //流程定义模板
	UserID     string `gorm:"column:user_id"`     //创建者ID
	Source     string `gorm:"column:source"`      //来源(引擎可能被多个系统、组件等使用，这里记下从哪个来源创建的流程
	CreateTime string `gorm:"column:create_time"` //创建时间
}
