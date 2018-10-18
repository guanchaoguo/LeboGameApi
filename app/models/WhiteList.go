package models

type WhiteList struct {
	Id              int32  `json:"id" xorm:"pk id"`
	Ip_info         string `json:"ip_info"`               //厅主IP列表,
	Agent_id        int    `json:"agent_id"`              //厅主ID,
	Agent_name      string `json:"agent_name"`            //厅主名称,
	Agent_domain    string `json:"agent_domain"`          //厅主域名,
	Agent_seckey    string `json:"agent_seckey"`          //秘钥,
	Seckey_exp_date string `json:"seckey_exp_date"`       //seckey最后有效时间,
	State           int    `json:"state"`                 //状态：1可用，0不可用,
	Agent_code      string `json:"agent_code"`            //代理商code，做为代理商玩家用户名前缀,
	Account_type    int    `json:"account_type" xorm:"-"` //代理商的账号类型
	Agent_id2       int    `json:"agent_id2" xorm:"-"`    //代理商id
	Connect_mode    int64     `json:"connect_mode" xorm:"-"`    //厅主扣费模式
}

func (*WhiteList) TableName() string {
	return "white_list"
}

func (WhiteList) GetWhiteList(id int) (*WhiteList, error) {
	var white_list WhiteList
	has, err := engineSlave.Where("agent_id = ? AND state = ?", id, 1).Get(&white_list)
	if err != nil || !has {
		return nil, err
	}
	return &white_list, nil
}

func (WhiteList) UpdateKey(agent_id int, seckey_exp_date string, agent_seckey string) (bool, error) {
	var white_list WhiteList
	white_list.Seckey_exp_date = seckey_exp_date
	white_list.Agent_seckey = agent_seckey
	res, err := engine.Where("agent_id=?", agent_id).Update(white_list)
	if res == 0 || err != nil {
		return false, err
	}
	return true, nil
}
