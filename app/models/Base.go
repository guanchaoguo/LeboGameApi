package models

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"golang-LeboGameApi/config"
	"gopkg.in/mgo.v2"
)

var engine ,engineSlave *xorm.Engine
var mongo *mgo.Session
var mongodbName string

func init() {
	//实例化master mysql
	initMysqlMaster()
	//实例化slave mysql
	initMysqlSlave()
	//实例化mongo
	mongodb()
}

func initMysqlMaster()  {
	mysqlConf := config.GetMysqlConf("master")
	engine, _ = xorm.NewEngine("mysql", mysqlConf.DB_USERNAME+":"+mysqlConf.DB_PASSWORD+"@tcp("+mysqlConf.DB_HOST+":"+mysqlConf.DB_PORT+")/"+mysqlConf.DB_DATABASE+"?charset="+mysqlConf.CHARSET)
	engine.SetMaxIdleConns(mysqlConf.SetMaxIdleConns)
	engine.SetMaxOpenConns(mysqlConf.SetMaxOpenConns)
}
func initMysqlSlave()  {
	mysqlConf := config.GetMysqlConf("slave")
	engineSlave, _ = xorm.NewEngine("mysql", mysqlConf.DB_USERNAME+":"+mysqlConf.DB_PASSWORD+"@tcp("+mysqlConf.DB_HOST+":"+mysqlConf.DB_PORT+")/"+mysqlConf.DB_DATABASE+"?charset="+mysqlConf.CHARSET)
	engineSlave.SetMaxIdleConns(mysqlConf.SetMaxIdleConns)
	engineSlave.SetMaxOpenConns(mysqlConf.SetMaxOpenConns)
}
//获取mongodb操作对象
func mongodb() bool{
	var err error
	url, dbName := config.GetMongodb()
	mongodbName = dbName
	//mongo, err = mgo.Dial(config.MONGO_USERNAME+":"+config.MONGO_PASSWORD+"@"+config.MONGO_HOST+":"+ config.MONGO_PORT+"/"+config.MONGO_DATABASE + "?connect="+ config.Connect+"&maxPoolSize="+config.MaxPoolSize)
	//mongo, err = mgo.Dial("mongodb://hhq163:bx123456@192.168.31.233:30000,192.168.31.234:30000,192.168.31.235:30000/live_game?connect=&maxPoolSize=")
	//mongo, err = mgo.DialWithTimeout(url,1*time.Second)
	mongo, err = mgo.Dial(url)
	if err != nil {
		fmt.Println("Connect to mongodb error", err)
		return false
	}
	mongo.SetMode(mgo.Eventual, true)
	//mongo.SetMode(mgo.Monotonic, true)
	return true
}

func GetMongodb() *mgo.Session {
	/*err := mongodb()
	if ! err  {
		return nil
	}*/
	err := mongo.Ping()
	if err != nil {
		return nil
	}
	return mongo.Copy()
}
