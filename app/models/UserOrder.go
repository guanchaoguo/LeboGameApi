package models

import (
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type UserOrder struct {
	User_name              	string        `json:"user_name"`
	Game_id               	int           `json:"game_id"`
	Game_hall_id          	int           `json:"game_hall_id"`
	Game_name             	string        `json:"game_name"`
	Bet_money           	float64        `json:"bet_money"`
	Bet_money_valid       	float64       `json:"bet_money_valid"`
	Round_no 				string       `json:"round_no"`
	Add_time           		time.Time        `json:"add_time"`
	//Id_                   bson.ObjectId `json:"_id,string" bson:"_id"`
}

func (this UserOrder) MarshalJSON() ([]byte, error) {
	type UserOrder2 UserOrder
	tmpUserOrder := struct {
		UserOrder2
		Add_time string `json:"add_time"`
	}{
		UserOrder2: (UserOrder2)(this),
		Add_time:           this.Add_time.Format("2006-01-02 15:04:05"),
	}
	return json.Marshal(tmpUserOrder)
}

func (UserOrder) GetListData(where bson.M, take int, sort string) ([]UserOrder, error) {
	session := GetMongodb()
	defer session.Close()
	c := session.DB(mongodbName).C("user_order")
	result := []UserOrder{}
	err := c.Find(where).Sort(sort).Limit(take).All(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
