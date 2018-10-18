package models

import (
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type ExceptionOrderLog struct {
	Id                bson.ObjectId `json:"_id,string" bson:"_id"`
	User_order_id     string        `json:"-"`
	Uid               int32         `json:"-" bson:"-"`
	User_name         string        `json:"user_name"`
	Agent_id          int64         `json:"-"`
	Agent_name        string        `json:"-"`
	Hall_id           int64         `json:"-"`
	Hall_name         string        `json:"hall_name"`
	Round_no          string        `json:"round_no"`
	Payout_win        float64       `json:"payout_win"`
	Before_user_money float64       `json:"-"`
	User_monry        float64       `json:"user_monry"`
	Bet_time          time.Time     `json:"bet_time,string"`
	Action_user       string        `json:"action_user"`
	Action_user_id    int64         `json:"action_user_id"`
	Action_passivity  string        `json:"-"`
	Add_time          time.Time     `json:"add_time"`
	Desc              string        `json:"desc"`
}

func (this ExceptionOrderLog) MarshalJSON() ([]byte, error) {
	type AliasExceptionOrderLog ExceptionOrderLog
	tmpUserChartInfo := struct {
		AliasExceptionOrderLog
		Bet_time string `json:"bet_time"`
		Add_time string `json:"add_time"`
	}{
		AliasExceptionOrderLog: (AliasExceptionOrderLog)(this),
		Bet_time:               this.Bet_time.Format("2006-01-02 15:04:05"),
		Add_time:               this.Add_time.Format("2006-01-02 15:04:05"),
	}
	return json.Marshal(tmpUserChartInfo)
}

/**
 * 通过时间获取异常注单数据
 * @param pkey string pkey
 * @return    *CashRecord 结果
 * @return    error 错误值
 */
func (ExceptionOrderLog) GetListData(pipe []bson.M, take int) ([]ExceptionOrderLog, error) {
	session := GetMongodb()
	defer session.Close()
	c := session.DB(mongodbName).C("exception_cash_log")
	result := []ExceptionOrderLog{}
	err := c.Pipe(pipe).All(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
