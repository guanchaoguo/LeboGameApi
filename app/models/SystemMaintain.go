package models

type SystemMintain struct {
	Id              int
	Sys_type        int    // 维护平台类型，0为平台，1为厅，默认为0,
	State           int    //系统是否开启维护，0为未开启，1为开启，默认为0
	Hall_id         string // 维护厅的id集合，用 ‘ , ’隔开'
	Start_date      string // 系统维护开始时间 0000-00-00 00:00:00
	End_date        string // 系统维护结束时间 0000-00-00 00:00:00
	Comtent         string //系统维护的公告内容
	Add_user        string // 开启系统维护的操作角色账号
	Add_date        string // 添加时间 0000-00-00 00:00:00
	User_start_date string // 用户添加的开始时间记录，和系统美东时间区别 0000-00-00 00:00:00
	User_end_date   string // 用户添加的结束时间记录，和系统美东时间区别 0000-00-00 00:00:00
}

func (*SystemMintain) TableName() string {
	return "system_maintain"
}

func (SystemMintain) GetInfo() (*SystemMintain, error) {
	var system_mintain SystemMintain
	has, err := engineSlave.Where("sys_type=? AND state = ?", 0, 1).Get(&system_mintain)
	if err != nil || !has {
		return nil, err
	}
	return &system_mintain, nil
}
