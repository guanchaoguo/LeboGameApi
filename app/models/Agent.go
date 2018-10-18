package models

import (
	"fmt"
)

type Agent struct {
	Id           int    `json:"id" xorm:"pk id"`
	User_name    string `json:"user_name"`
	Parent_id    int    `json:"parent_id"`
	Agent_code   string `json:"agent_code"`
	AccountState int    `json:"account_state"`
	AccountType  int    `json:"account_type"`
	Sub_user     int    `json:"sub_user"`
	Connect_mode     int64    `json:"connect_mode"`
}

func (*Agent) TableName() string {
	return "lb_agent_user"
}

func (Agent) GetAgentInfo(agent_name string) (*Agent, error) {
	var agent Agent
	has, err := engineSlave.Where("user_name=? and account_state =? and grade_id=? ", agent_name, 1, 2).Get(&agent)
	if err != nil || !has {
		return nil, err
	}
	return &agent, nil
}

func (Agent) GetHallInfo(id int) (*Agent, error) {
	var agent Agent
	has, err := engineSlave.Where("id = ? AND account_state =? AND grade_id = ? AND is_hall_sub = ? ", id, 1, 1, 0).Get(&agent)
	if err != nil || !has {
		return nil, err
	}
	return &agent, nil
}

/**
 * 判断是否为联调账号
 * @param agent_name string 代理商名称
 * @return 			 bool 返回结果
 */
func (Agent) IsDebugAccount(agent_name string) bool {
	var agent Agent
	has, err := engineSlave.Where("user_name=? and account_state =? and grade_id=?  and account_type=?", agent_name, 1, 2, 3).Get(&agent)
	if err != nil || !has {
		return false
	}
	return true
}

/**
 * 更新代理商、厅主 玩家数
 * @param id int 代理商id、厅主id
 * @return    bool
 */
func (Agent) SubUserNum(id int) bool {
	var agent Agent
	has, err := engineSlave.Where("id=?", id).Get(&agent)
	if err != nil || !has {
		fmt.Println(err)
		return false
	}
	agent.Sub_user += 1
	res, err := engine.Where("id=?", id).Update(&agent)
	if res == 0 || err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (Agent) GetAgentTest(id int) (*Agent, error) {
	user := &Agent{Id: id}
	_, err := engineSlave.Get(user)
	return user, err
}
