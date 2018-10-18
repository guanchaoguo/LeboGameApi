package models

import (
	"github.com/tidwall/gjson"
	"strconv"
	"time"
)

/**
 * 代理商投注派彩统计
 */
type StatisCashAgent struct {
	Id                 int     `json:"id" xorm:"pk"`
	Add_date           string  `json:"add_date"`
	Day_year           int     `json:"day_year"`
	Day_month          int     `json:"day_month"`
	Day_day            int     `json:"day_day"`
	Hall_id            int     `json:"hall_id"`
	Hall_name          string  `json:"hall_name"`
	Agent_id           int     `json:"agent_id"`
	Agent_name         string  `json:"agent_name"`
	Total_score_record float64 `json:"total_score_record"`
}

func (*StatisCashAgent) TableName() string {
	return "statis_cash_agent"
}

/**
 * 统计代理商玩家的充值
 */
func (StatisCashAgent) TotalScoreRecord(agent_name string, whitelist string, money float64) bool {

	account_type := gjson.Get(whitelist, "account_type").Int()
	agent_id, _ := strconv.Atoi(gjson.Get(whitelist, "agent_id2").String())
	hall_id, _ := strconv.Atoi(gjson.Get(whitelist, "agent_id").String())
	hall_name := gjson.Get(whitelist, "agent_name").String()

	//不是正常账号，账号不正常的不统计
	if account_type != 1 {
		return false
	}

	var cash_agent StatisCashAgent

	now_time := time.Now().Format("2006-01-02")
	has, err := engineSlave.Where("add_date=? and agent_id=?", now_time, agent_id).Get(&cash_agent)
	if !has || err != nil {
		//新增
		var insert_data StatisCashAgent
		insert_data.Add_date = now_time
		insert_data.Day_year = time.Now().Year()
		insert_data.Day_month = int(time.Now().Month())
		insert_data.Day_day = time.Now().Day()
		insert_data.Agent_id = agent_id
		insert_data.Hall_id = hall_id
		insert_data.Agent_name = agent_name
		insert_data.Hall_name = hall_name
		insert_data.Total_score_record = money
		engine.Insert(&insert_data)
		return true
	} else {
		//更新
		//cash_agent.Total_score_record += money
		//res, err := engine.Where("add_date=? and agent_id=?",now_time, agent_id).Update(&cash_agent)
		data := map[string]interface{}{
			"total_score_record": cash_agent.Total_score_record + money,
		}
		engine.Table("statis_cash_agent").Where("id=?", cash_agent.Id).Update(data)

		return true
	}
}
