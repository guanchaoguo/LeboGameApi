package helper

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"golang-LeboGameApi/config"
	"io"
	"math/big"
	rand2 "math/rand"
	"strconv"
	"strings"
	"time"
)

var RedisClientDefault *redis.Pool
var RedisClientMonitor *redis.Pool
var RedisClientAccount *redis.Pool

func init()  {
	//初始化redis连接池
	RedisConnPoll()
}

func MakePassword(len int) string {
	res := ""
	chars := [...]string{
		"a", "b", "c", "d", "e", "f", "g", "h",
		"i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s",
		"t", "u", "v", "w", "x", "y", "z", "A", "B", "C", "D",
		"E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O",
		"P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	}

	for i := 0; i < len; i++ {
		index := RandInt64(0, 61)
		res += chars[index]
	}
	return res
}

/**
 * 获取指定长度的随机字符串
 * @param len int 字符串长度
 * @return    string 返回一个随机字符串
 */
func Randomkeys(len int) string {
	res := ""
	str := "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLOMNOPQRSTUVWXYZ"
	for i := 0; i < len; i++ {
		index := RandInt64(0, 61)
		res += fmt.Sprintf("%c", str[index])
	}

	return res
}

/**
 * 生成单号sn
 */
func CreateOrderSn(prefix string) string {
	if prefix == "" {
		prefix = "L"
	}
	code := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}

	year := time.Now().Year()
	index := int(year) - 2017
	if index >= len(code) {
		index = 0
	}

	month := int(time.Now().Month())
	mCode := strings.ToUpper(fmt.Sprintf("%01X", month))

	day := strconv.Itoa(time.Now().Day())
	microTime := strconv.FormatInt(time.Now().UnixNano()/1e3, 10)

	r := rand2.New(rand2.NewSource(time.Now().UnixNano()))
	spri := fmt.Sprintf("%02d", r.Intn(100))

	orderSn := prefix + code[index] + mCode + day + microTime[5:15] + spri
	return orderSn
}

/**
 * 对字符串进行MD5哈希
 * @param data string 要加密的字符串
 */
func MD5(data string) string {
	t := md5.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}

/**
 * 对字符串进行SHA1哈希
 * @param data string 要加密的字符串
 */
func SHA1(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}

/**
 * 参数加密token检测
 * @param token  string  参数token
 * @param slice param  加密字段切片，注意数据的下标顺序要和字段拼接加密顺序一致
 * @return bool
 */
func CheckEequest(token string, param []string) bool {
	if param == nil {
		return false
	}
	var str string
	for key, item := range param {
		if len(param) >= 1 && (key+1) >= len(param) {
			str += item
		} else {
			str += item + "|"
		}
	}
	//fmt.Println(SHA1(str))

	if token != SHA1(str) {
		return false
	}
	return true
}

/**
 *生成随机数
 *@param  min: 最小值
 * @param max: 最大值
 * @return int64: 生成的随机数
 */
func RandInt64(min, max int64) int64 {
	maxBigInt := big.NewInt(max)
	i, _ := rand.Int(rand.Reader, maxBigInt)
	if i.Int64() < min {
		RandInt64(min, max)
	}
	return i.Int64()
}

/**
 * 写日志文件
 */
/*func Self_logger(myerr interface{}) {
	logfile, err := os.OpenFile(newLogFile(), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0)
	if err != nil {
		fmt.Printf("%s\r\n", err.Error())
		os.Exit(-1)
	}
	defer logfile.Close()
	logger := log.New(logfile, "\r\n", log.Ldate|log.Ltime|log.Llongfile)
	logger.Println(myerr)

}*/

/*func todayFilename() string {
	today := time.Now().Format("2006-01-02")
	return today + ".log"
}

func newLogFile() string {
	losPath,_ := os.Getwd()
	filename := todayFilename()
	path:= losPath + "/logs/access/" + filename
	return path
}*/

//获取redis操作对象
func GetRedis(connect string) redis.Conn {
	redisConf := config.GetRedis(connect)
	c, err := redis.Dial("tcp", redisConf.REDIS_HOST+":"+redisConf.REDIS_PORT, redis.DialPassword(redisConf.REDIS_PASSWORD))
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return nil
	}
	return c
}

//redis连接池
func RedisConnPoll()  {

	//默认业务池
	redisConf := config.GetRedis("default")
	RedisClientDefault = &redis.Pool{
		MaxIdle:10,//空闲链接
		MaxActive:100,//最大激活连接数
		IdleTimeout:180*time.Second,//操作该时间后，空闲的链接将会被关闭
		Dial: func() (redis.Conn, error) {
			c,err := redis.Dial("tcp", redisConf.REDIS_HOST+":"+redisConf.REDIS_PORT, redis.DialPassword(redisConf.REDIS_PASSWORD))
			if err != nil{
				return nil,err
			}
			return c,nil
		},
	}

	//monitor业务池
	redisConfMonitor := config.GetRedis("monitor")
	RedisClientMonitor = &redis.Pool{
		MaxIdle:10,//空闲链接
		MaxActive:100,//最大激活连接数
		IdleTimeout:180*time.Second,//操作该时间后，空闲的链接将会被关闭
		Dial: func() (redis.Conn, error) {
			c,err := redis.Dial("tcp", redisConfMonitor.REDIS_HOST+":"+redisConfMonitor.REDIS_PORT, redis.DialPassword(redisConfMonitor.REDIS_PASSWORD))
			if err != nil{
				return nil,err
			}
			return c,nil
		},
	}

	//monitor业务池
	redisConfAccount := config.GetRedis("account")
	RedisClientAccount = &redis.Pool{
		MaxIdle:10,//空闲链接
		MaxActive:100,//最大激活连接数
		IdleTimeout:180*time.Second,//操作该时间后，空闲的链接将会被关闭
		Dial: func() (redis.Conn, error) {
			c,err := redis.Dial("tcp", redisConfAccount.REDIS_HOST+":"+redisConfAccount.REDIS_PORT, redis.DialPassword(redisConfAccount.REDIS_PASSWORD))
			if err != nil{
				return nil,err
			}
			return c,nil
		},
	}



}

/**
 * 添加api调用和会员登录日志信息
 */

func addApiLog(data map[string]string) {
	if len(data) == 0 {
		return
	}

}

/**
 *  随机打乱一个字符串顺序
 */

func Shuffle_str(str string) string {
	if len(str) == 0 {
		return ""
	}

	var strMap = make(map[int]string)
	for i, v := range str {
		strMap[i] = string(v)
	}

	var shuffle_str string
	for _, v := range strMap {
		shuffle_str += v
	}
	return shuffle_str
}
