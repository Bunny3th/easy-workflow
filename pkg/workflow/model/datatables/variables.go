package datatables

type Variable struct {
	id      int `gorm:"column:id"`
	proc_id int `gorm:"column:id"`   //'流程ID'
	key     string // '变量key'
	value   string //'变量value'

}
