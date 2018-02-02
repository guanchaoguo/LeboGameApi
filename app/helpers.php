<?php
/**
 * Created by PhpStorm.
 * User: liangxz@szljfkj.com
 * Date: 2017/3/14
 * Time: 13:55
 * 帮助函数
 */
use bootstrap\Configuration;
use bootstrap\Input;
use bootstrap\Bootstrap;
use lib\db\DB;
/**
 * 创建订单号
 * @param string $prefix
 * @return string
 */
function createOrderSn($prefix = "L")
{
    $yCode = array('A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J');
    return strtoupper($prefix).$yCode[intval(date('Y')) - 2017] . strtoupper(dechex(date('m'))) . date('d') . substr(time(), -5) . substr(microtime(), 2, 5) . sprintf('%02d', rand(0, 99));
}

/**
 * 返回json数据
 * @param array $data
     [
        'code' => 0,
        'text' => '操作成功',
        'result' => '',
    ]
 */
function returnJson($data=[])
{
     header('Content-type: text/json');
     echo json_encode($data);die;
}


if (! function_exists('config')) {
    /**
     * @param  array|string  $key
     * @return mixed
     */
    function config($key)
    {
        if (is_null($key)) {
            return $key;
        }
        $key_arr = explode('.', $key);
       return Configuration::env($key_arr[1],$key_arr[0]);
    }
}

if (! function_exists('get_real_ip')) {
    /**
     * 获取ip地址
     * @return string
     */
    function get_real_ip()
    {
        static $realip;
        if(isset($_SERVER)){
            if(isset($_SERVER['HTTP_X_FORWARDED_FOR'])){
                $realip=$_SERVER['HTTP_X_FORWARDED_FOR'];
            }else if(isset($_SERVER['HTTP_CLIENT_IP'])){
                $realip=$_SERVER['HTTP_CLIENT_IP'];
            }else{
                $realip=$_SERVER['REMOTE_ADDR'];
            }
        }else{
            if(getenv('HTTP_X_FORWARDED_FOR')){
                $realip=getenv('HTTP_X_FORWARDED_FOR');
            }else if(getenv('HTTP_CLIENT_IP')){
                $realip=getenv('HTTP_CLIENT_IP');
            }else{
                $realip=getenv('REMOTE_ADDR');
            }
        }
        return $realip;
    }
}

/**
 * @param $token  Request  参数token
 * @param array $param array 加密字段数组，注意数据的下标顺序要和字段拼接加密顺序一致
 * @return bool
 */
function checkEequest($token,$param = [],$test=0)
{
    if(!is_array($param))
        return false;

    $str = "";
    foreach ($param as $k=>$v)
    {
        if(count($param) >= 1 && $k+1>= count($param))
        {
            $str .= "$v";
        }
        else
        {
            $str .= "$v"."|";
        }
    }

    if($test){
        var_export(hash(config('common.SECURITY_KEY_ENCRYPT'),$str));
        var_export($str);die;
    }
    //进行加密比对
    if($token != hash(config('common.SECURITY_KEY_ENCRYPT'),$str))
    {
        return  false;
    }
    return true;
}

/**
 * 根据agent获取对应的SecurityKey
 * @param $agent
 */
function getSecurityKey($agent_name){
    if(empty($agent_name))
        return false;

    // 获取缓存数据中的数据是否命中名单信息
    $agentKeyName = 'agentWhitelist';
    if (!$ips = Bootstrap::Redis()->hGet($agentKeyName,$agent_name)) {
        // 通过代理商获取厅主信息信息
        $DB = new DB();
        if(!$agentInfo = getHallByAgent($agent_name)){
            return returnJson([
                'code' => 9004,//代理所属的厅主不存在
                'text' => config('errorcode.9004'),
                'result' => '',
            ]);
        }

        // 未命中缓存获取数据库白名单信息
        $whites = $DB->row("SELECT * FROM white_list WHERE agent_id = :agent_id AND state = :state", ['agent_id' => $agentInfo['parent_id'], 'state' =>1]);
        $ips = json_encode($whites);
    }

    $ips =json_decode($ips, true);
    return $ips['agent_seckey'];
}

/*
 * 通过代理获取厅主信息
 */
function getHallByAgent($agent_name){
    $DB = new DB();
    $where = [
        'user_name' => $agent_name,
        'account_state' => 1,//正常状态
        'grade_id' => 2,//二级代理商
    ];
    $agent = $DB->row("SELECT * FROM lb_agent_user WHERE user_name = :user_name AND account_state = :account_state AND grade_id = :grade_id", $where);
    if(!$agent)   return false;

    return $agent;
}

//返回数据验证错误代码
function returnError()
{
    return returnJson([
        'code'  => 9002,
        'text'  => config('errorcode.9002'),
        'result'    => ''
    ]);
}

function make_password($length=6){
    // 密码字符集，可任意添加你需要的字符 
    $chars=array(
        'a','b','c','d','e','f','g','h',
        'i','j','k','l','m','n','o','p','q','r','s',
        't','u','v','w','x','y','z','A','B','C','D',
        'E','F','G','H','I','J','K','L','M','N','O',
        'P','Q','R','S','T','U','V','W','X','Y','Z',
        '0','1','2','3','4','5','6','7','8','9'
    );
    $keys = array_rand($chars,$length);
    $password = '';
    for ($i=0;$i<$length;$i++){
        $password .= $chars[$keys[$i]];
    }
    return $password;
}

/**
 * 获取随机数
 * @param int $length
 * @return string
 */
function randomkeys(int $length) {
    $returnStr='';
    $pattern = '1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLOMNOPQRSTUVWXYZ';
    for($i = 0; $i < $length; $i ++) {
        $returnStr .= $pattern {mt_rand ( 0, 61 )}; //生成php随机数
    }
    return str_shuffle($returnStr);
}
/**
 * 添加api调用和会员登录日志信息
 */

function addApiLog($data = [])
{
    if(empty($data)) return false;
    
    // 将日志写入消息队列    
    $mqModel = new App\Model\MqProducer;
    $mqModel-> publishMsg(json_encode($data));
}

/**
 *  判断是否为联调账号
 * @param array $data
 * @return bool
 */
function is_debugAccount($data = []){
    if(empty($data)) return false;

    $DB = new DB();
    $data['agent'] = Input::post('agent');
    $where = [
        'user_name' => $data['agent'],
        'account_state' => 1,//正常账号
        'account_type' => 3,//联调账号
    ];

    // 查询是否为联调账号
    $agent_ = $DB->row("SELECT id,parent_id FROM lb_agent_user WHERE user_name = :user_name AND account_state = :account_state AND account_type = :account_type", $where);
    if(empty($agent_)){
        return false;
    }else{
        return true;
    }
}

/**
 * 联调接口调用总次数统计
 */
function ApiStatistics($data = []){
    // 查询是否为联调账号
    if (!is_debugAccount($data)) return false;

    // 判断该联调代理数据是否存在 否则更新操作
    $db = \bootstrap\Bootstrap::MongoDB()->selectCollection('live_game', 'api_statistics_log');

    // 查询是否存在该厅住的联调信息
    $data['agent'] = Input::post('agent');
    $result = $db->find(['apiName' => $data['apiName'], 'agent' => $data['agent']])->toArray();
    if (empty($result)) {
        // 第一次统计数据
        $db->insertOne([
            'apiName' => $data['apiName'],
            'agent' => $data['agent'],
            'status' => 0,// 是否联调通过
            'succeds' => 0,// 成功次数
            'sum' => 1,// 联调总次数
        ]);
    } else {
        // 更新次数
        $filter = ['apiName' => $data['apiName'], 'agent' => $data['agent']];
        $update = ['$inc' => ['sum' => 1]];
        $db->updateOne($filter, $update);
    }

    return true;
}

    /**
    * 联调接口调用成功次数统计
    */
    function ApiSucceds($data = []){
        // 查询是否为联调账号
        if(!is_debugAccount($data)) return false;

        // 修改联调账号成功次数
        $db = \bootstrap\Bootstrap::MongoDB()->selectCollection('live_game','api_statistics_log');
        $data['agent'] = Input::post('agent');
        $filter = ['apiName'=>$data['apiName'], 'agent'=>$data['agent']];
        $update = [
            '$inc'=>['succeds'=>1],
            '$set'=>['status'=>1]
        ];
        $db->updateOne( $filter,$update);
    }


