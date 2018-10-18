package models

import (
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type UserChartInfo struct {
	Round_no              string        `json:"round_no"`
	Game_id               int           `json:"game_id"`
	Game_hall_id          int           `json:"game_hall_id"`
	Game_name             string        `json:"game_name"`
	Server_name           string        `json:"server_name"`
	Total_bet_score       float64       `json:"total_bet_score"`
	Valid_bet_score_total float64       `json:"valid_bet_score_total"`
	Game_result           string        `json:"game_result"`
	Total_win_score       float64       `json:"total_win_score"`
	End_time              time.Time     `json:"-"`
	User_name             string        `json:"user_name"`
	Is_mark               int           `json:"is_mark"`
	DwRound               int           `json:"dwRound" bson:"dwRound"`
	Remark                string        `json:"remark"`
	Id_                   bson.ObjectId `json:"_id,string" bson:"_id"`
	Game_period			  string		`json:"game_period"`
}

func (this UserChartInfo) MarshalJSON() ([]byte, error) {
	type AliasUserChartInfo UserChartInfo
	tmpUserChartInfo := struct {
		AliasUserChartInfo
		End_time string `json:"start_time"`
	}{
		AliasUserChartInfo: (AliasUserChartInfo)(this),
		End_time:           this.End_time.Format("2006-01-02 15:04:05"),
	}
	return json.Marshal(tmpUserChartInfo)
}

/**
 * 通过时间获取注单数据
 * @param pkey string pkey
 * @return    *CashRecord 结果
 * @return    error 错误值
 */
func (UserChartInfo) GetListData(findWhere bson.M, take int, sort string) ([]UserChartInfo, error) {
	session := GetMongodb()
	defer session.Close()

	c := session.DB(mongodbName).C("user_chart_info")
	result := []UserChartInfo{}
	err := c.Find(findWhere).Sort(sort).Limit(take).All(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
