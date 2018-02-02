<?php
/**
 * Created by PhpStorm.
 * User: chensongjian
 * Date: 2017/3/20
 * Time: 13:35
 */

namespace App\controllers;

use bootstrap\Bootstrap;
use bootstrap\Input;
use lib\db\DB;

class IpLimit
{

    //是否ip限制
    const ip_limit = true;
    public function __construct()
    {

    }

    /**
     * IP限制
     * @return mixed
     * 一般来说代理商的IP就是厅主的IP，如果代理商和厅主的IP不一样，
     * 就必须要在系统的白名单中把该代理商的IP添加入所属的厅主白名单中
     */
    public static function init()
    {
        self::getAjaxAgent();// 为正式环境 赋值 代理厅主

        //免费试玩 检查试玩代理白名单信息
        if(Input::post('account_type') == 2 ) {
            if(!self::testPlayer()){
                return returnJson([
                    'code' => 1005,//创建会员错误
                    'text' => config('errorcode.1003'),
                    'result' => '',
                ]);
            }
        }

        $agent_name = Input::post('agent');
        if (empty($agent_name)) {
            return returnJson([
                'code' => 9002,//请求参数错误
                'text' => config('errorcode.9002'),
                'result' => ''
            ]);

        }

        if (self::ip_limit) {

            // 获取缓存中白名单信息
            $agentKeyName = 'agentWhitelist'; $DB = new DB();

            if (!$ips = Bootstrap::Redis()->hGet($agentKeyName,$agent_name)) {
                if(!$agentInfo = getHallByAgent($agent_name)){
                    return returnJson([
                        'code' => 9004,//代理所属的厅主不存在
                        'text' => config('errorcode.9004'),
                        'result' => '',
                    ]);
                }
                // 未命中缓存获取数据库白名单信息
                $whites = $DB->row("SELECT * FROM white_list WHERE agent_id = :agent_id AND state = :state", ['agent_id' => $agentInfo['parent_id'], 'state' =>1]);
                // 类型转换
                $whites['id'] = (int)$whites['id'];
                $whites['agent_id'] = (int)$whites['agent_id'];
                $whites['state'] = (int)$whites['state'];
                // 额外字段
                $whites['agent_code'] = $agentInfo['agent_code'];
                $whites['account_type'] = (int)$agentInfo['account_type'];
                $whites['agent_id2'] = (int)$agentInfo['id'];
                $ips = json_encode($whites);
                Bootstrap::Redis()->hSet($agentKeyName,$agent_name,$ips);
            }

        }

        //直接获取信息
        if ($ips) {
            $ips = json_decode($ips, true);
            $remote_ip = get_real_ip();
            if ($ips['ip_info'] != '*' && !strstr($ips['ip_info'], $remote_ip)) {
                return returnJson([
                    'code' => 9003,//IP限制
                    'text' => config('errorcode.9003'),
                    'result' => ''
                ]);
            }
        }

    }


    /**
     * 判断系统是否在维护当中
     * @param string $username
     * @param string $agent
     * @return bool|void
     *
     */
    public  static function checkIsMaintain($username = "", $agent="")
    {
        // 获取缓存的是否命中的agent_code
        $agentKeyName = 'agentWhitelist'; $DB = new DB();
        $ips = Bootstrap::Redis()->hGet($agentKeyName,$agent);
        $ips = (array)json_decode($ips,1);
        if ( !$ips || empty($ips['agent_code'])) {

            // 未命中缓存则直接查询数据库
            if(!$agentInfo = getHallByAgent($agent)){
                return returnJson([
                    'code' => 9004,//代理所属的厅主不存在
                    'text' => config('errorcode.9004'),
                    'result' => '',
                ]);
            }
            $ips['agent_code']  = $agentInfo['agent_code'];
        }

        $decry_username = Crypto::decrypt($ips['agent_code'].$username);
        $lb_user = $DB->row("SELECT * FROM lb_user WHERE user_name = :user_name", ['user_name' => $decry_username]);

        //管理员身份所有时间段都能进行登录
        if($lb_user['user_rank'] == 2) return false;

        //获取系统维护信息
        $info = $DB->row("select * from system_maintain where sys_type=0  and state=1");

        if( isset($info) && !$info = $info)
            return false;//没有维护信息证明不在维护，所以返回假

        //判断当前时间是否在系统维护时间内
        if(strtotime($info['start_date']) <= time() && strtotime($info['end_date']) >= time() )
        {
            //系统正在维护当中
            return returnJson([
                'code' => 9001,
                'text' => config('errorcode.9001'),
                'result' =>''
            ]);
        }
    }

    /**
     * 判断正式环境的使用代理
     * @return string
     */
    public static  function getAjaxAgent():bool
    {
        // 判断有效的来源 线上环境则返回true
        if(  'ONLINE' == SETTING ){
            if( $_SERVER['SERVER_NAME'] == 'lgapishow.lggame.co') {
                $_POST['agent'] = $_REQUEST['agent'] = 'dlebo01';
                return true;
            }
            return false;
        }
        return false;
    }

    /**
     * 单独校验试玩玩家厅主白名单
     * @return bool
     */
    private static function testPlayer():bool
    {
        // 获取缓存中白名单信息
        $agentKeyName = 'agentWhitelist'; $DB = new DB();$agent_name ='agent_test';
        if (!$ips = Bootstrap::Redis()->hGet($agentKeyName,$agent_name)) {
            if(!$agentInfo = getHallByAgent($agent_name)){
                return false;// 试玩代理不存在
            }
            // 未命中缓存获取数据库白名单信息
            $whites = $DB->row("SELECT * FROM white_list WHERE agent_id = :agent_id AND state = :state", ['agent_id' => $agentInfo['parent_id'], 'state' =>1]);

            // 类型转换
            $whites['id'] = (int)$whites['id'];
            $whites['agent_id'] = (int)$whites['agent_id'];
            $whites['state'] = (int)$whites['state'];
            // 额外字段
            $whites['agent_code'] = $agentInfo['agent_code'];
            $whites['account_type'] = (int)$agentInfo['account_type'];
            $whites['agent_id2'] = (int)$agentInfo['id'];
            $ips = json_encode($whites);
            Bootstrap::Redis()->hSet($agentKeyName,$agent_name,$ips);
        }

        //直接获取信息
        if ($ips) {
            $ips = json_decode($ips, true);
            $remote_ip = get_real_ip();
            if ($ips['ip_info'] != '*' && !strstr($ips['ip_info'], $remote_ip)) {
                return  false;// 试玩厅主白名单不合法
            }
        }
        return true;
    }

}