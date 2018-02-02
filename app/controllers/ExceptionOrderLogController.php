<?php
/**
 * Created by PhpStorm.
 * User: liangxz@szljfkj.com
 * Date: 2017/3/21
 * Time: 10:02
 * 代理商获取注单信息
 */
namespace App\controllers;

use bootstrap\Bootstrap;
use bootstrap\Input;
use lib\db\DB;

class ExceptionOrderLogController
{

    public function __construct()
    {
        IpLimit::init();//白名单过滤
        SecurityKeyMiddleware::handle();//SecurityKey过滤

    }

    /**
     * @SWG\Post(
     *   path="/exceptionOrderLog",
     *   tags={"api"},
     *   summary="获取异常注单日志信息",
     *   description="
     * 时间段获取代理商异常游戏注单日志信息
     * PS:需要注意的是：开始时间和结束时间跨度甭能超过一个星期
     *
     * 接口信息返回说明：
     * code： 0 成功；
     * text ：'接口错误信息描述'；
     *result: {
     *  data: [
     *       {
     *           user_order_id: A333e138eee37b244ec4a2, //注单ID
     *           user_name: a9TEST717929,//玩家名
     *           agnet_name: agnet_test,// 代理名
     *           hall_name: csj,//厅主名
     *           round_no: 71a45f7196da5d8e,// 局ID
     *           payout_win: 500,//派彩金额
     *           user_money: 500,// 代理余额  当取消注单扣除下注金额为负数的时候 需要存款操作 当取消注单返还下注金额的时候 代理可以进行取款操作
     *           bet_time: 2017-06-19 05:54:52 // 下注时间
     *           desc: 取消异常派彩,//备注
     *           action_user: a9TEST717929,//操作者
     *           add_time: 2017-06-19 06:29:58//操作时间
     *       }
     *   ]
     *}
     *   
     *     ",
     *   operationId="index",
     *   @SWG\Parameter(
     *     in="formData",
     *     name="agent",
     *     type="string",
     *     description="代理商用户名",
     *     required=true,
     *     default="h8888"
     *   ),
     *   @SWG\Parameter(
     *     in="formData",
     *     name="start_date",
     *     type="string",
     *     format="date",
     *     description="开始时间",
     *     required=true,
     *     default="2017-06-18 08:21:02"
     *   ),
     *   @SWG\Parameter(
     *     in="formData",
     *     name="end_date",
     *     type="string",
     *     format="date",
     *     description="结束时间",
     *     required=true,
     *     default="2017-06-24 08:21:02"
     *   ),
     *   @SWG\Parameter(
     *     in="formData",
     *     name="token",
     *     type="string",
     *     description="SHA1('securityKey|start_date|end_date|agent')",
     *     required=true,
     *     default="138637fc5594d70bb8f63daa8c71be0d9a47a1c1"
     *   ),
     *   @SWG\Response(response="200", description="Success")
     * )
     */
    public function index()
    {
        // 日志捕获
        $apiLog['start_time'] = time();
        $apiLog['agent'] = Input::post('agent');
        $apiLog['postData'] = json_encode($_REQUEST);
        $apiLog['apiName'] = '时间段获取异常注单信息';
        $apiLog['ip_info'] = get_real_ip();

        $agent = Input::post('agent');
        $start_date = Input::post('start_date');
        $end_date = Input::post('end_date');
        $token = Input::post('token');

        // 联调状态统计 只要有联调请求则统计数据
        $debugging = false;
        if(ApiStatistics($apiLog))  $debugging = true;

        //var_export(sha1(getSecurityKey($agent)."|".$start_date."|".$end_date."|".$agent));die;
        //判断调用频率间隔时间
        if($feedError = $this->getListFeed($agent))
        {
            $apiLog['code'] = 9008;
            $apiLog['text'] = config('errorcode.9008');
            $apiLog['result'] = '';
            $apiLog['end_time'] = new \MongoDB\BSON\UTCDateTime(time() * 1000);
            addApiLog($apiLog);//日志记录
       
            return $feedError;
        }

        //数据验证错误
        if(!$this->makeDateValidate())
        {
            $apiLog['code'] = 9002;
            $apiLog['text'] = config('errorcode.9002');
            $apiLog['result'] = '';
            $apiLog['end_time'] = new \MongoDB\BSON\UTCDateTime(time() * 1000);
            addApiLog($apiLog);//日志记录

 
            return returnError();
        }

        //判断系统是否在维护当中
        if( $maintain = IpLimit::checkIsMaintain(Input::post('username'), Input::post('agent')) )
        {
            return $maintain;
        }

        //时间跨度验证
        if(!$this->checkDatesSpacing($start_date,$end_date))
        {
            $apiLog['code'] = 2001;
            $apiLog['text'] = config('errorcode.2001');
            $apiLog['result'] = '';
            $apiLog['end_time'] = new \MongoDB\BSON\UTCDateTime(time() * 1000);
            addApiLog($apiLog);//日志记录

            return returnJson([
                'code'      => 2001,
                'text'      => config('errorcode.2001'),
                'result'    => ''
            ]);
        }

        //token验证，判断数据是否被篡改
        $param = [getSecurityKey($agent),$start_date,$end_date,$agent];
        if(!checkEequest($token,$param) || !$this->checkStartAndEndDate($start_date,$end_date))
        {
            $apiLog['code'] = 9002;
            $apiLog['text'] = config('errorcode.9002');
            $apiLog['result'] = '';
            $apiLog['end_time'] = new \MongoDB\BSON\UTCDateTime(time() * 1000);
            addApiLog($apiLog);//日志记录

            echo 3;
            return returnError();
        }

        //$agent = 'anchen2';
        //根据代理商获取代理商ID
//        $agentInfo = Bootstrap::DB()->get("lb_agent_user","*",['user_name'=>$agent]);
        $DB = new DB();
        $agentInfo = $DB->row("SELECT * FROM lb_agent_user WHERE user_name = :user_name", ['user_name'=>$agent]);
        if(!$agentInfo)
        {
            $apiLog['code'] = 9004;
            $apiLog['text'] = config('errorcode.9004');
            $apiLog['result'] = '';
            $apiLog['end_time'] = new \MongoDB\BSON\UTCDateTime(time() * 1000);
            addApiLog($apiLog);//日志记录

            return returnJson([
                'code'  => 9004,
                'text'  => config('errorcode.9004'),
                'result'   => ''
            ]);
        }

        //验证通过进行数据获取操作
        $result = $this->getOrderList($agentInfo['id'],0,$start_date,$end_date);

        //数据为空，返回空数据提示
        if(!$result)
        {
            $apiLog['code'] = 9009;
            $apiLog['text'] = config('errorcode.9009');
            $apiLog['result'] = '';
            $apiLog['end_time'] = new \MongoDB\BSON\UTCDateTime(time() * 1000);
            addApiLog($apiLog);//日志记录

            return returnJson([
                'code'      => 9009,
                'text'      => config('errorcode.9009'),
                'result'    => ''
            ]);
        }

        $res = [];

        //时间格式转换
        foreach ($result as $key=>$val)
        {

            $bet_date = date("Y-m-d H:i:s",ceil($val['bet_time']->__toString()/1000));
            $start_date = date("Y-m-d H:i:s",ceil($val['add_time']->__toString()/1000));
            $result[$key]['bet_time'] = $bet_date;
            $result[$key]['add_time'] = $start_date;
        }

        $res['data'] = $result;

        $apiLog['code'] = 0;
        $apiLog['text'] = config('errorcode.9999');
        $apiLog['result'] = json_encode($result);
        $apiLog['end_time'] = new \MongoDB\BSON\UTCDateTime(time() * 1000);
        addApiLog($apiLog);//日志记录
        if($debugging) ApiSucceds($apiLog);//联调账号联调成功次数统计

        //正常返回数据
        return returnJson([
            'code'      => 0,
            'text'      => config('errorcode.9999'),
            'result'    => $res
        ]);
    }

    //验证开始时间不能大于结束时间
    private function checkStartAndEndDate($start_date,$end_date)
    {
        if(strtotime($start_date) >= strtotime($end_date))
            return false;
        return true;
    }

        //数据验证
    private function makeDateValidate()
    {
        $agent = Input::post('agent');
        $start_date = Input::post('start_date');
        $end_date = Input::post('end_date');
        $token = Input::post('token');
        $is_start_time = strtotime($start_date) ? strtotime($start_date) : false;
        $is_end_time = strtotime($end_date) ? strtotime($end_date) : false;
        if(empty($agent) || empty($start_date) || empty($end_date) || empty($token) || !$is_start_time || !$is_end_time)
        {
            return false;
        }
        return true;
    }

    /**
     * 验证两个时间跨度
     * @param $startDate
     * @param $endDate
     * @return bool
     */
    private function checkDatesSpacing($startDate,$endDate)
    {
        $dateMargin = strtotime($endDate) - strtotime($startDate);
        if(($dateMargin/86400) > (int)config('common.MAX_TIME_SPAN'))
        {
            return false;
        }
        return true;
    }


    /**
     * 接口限定操作频率操作
     * @param $agent
     * 每次获取数据间隔为3秒,该时间可以在配置文件.env中修改 GET_ORDER_LIST_SPEED 参数对应的值
     */
    private function getListFeed($agent)
    {
        $keyName = $agent.config('common.AGENT_ORDER_LIST');
        $lastDate = Bootstrap::Redis()->get($keyName);

        //判断是否在有效时间间隔内
        if($lastDate && time()-$lastDate < config('common.GET_ORDER_LIST_SPEED'))
        {
            return returnJson([
                'code'      => 9008,
                'text'      => config('errorcode.9008'),
                'result'    => ''
            ]);
        }

        //符合速率则刷新本次调用时间
        Bootstrap::Redis()->set($keyName,time());

        return;
    }

    //获取数据操作
    private function getOrderList($agent_id,$deadline=0,$startDate="",$endDate="")
    {
        //如果代理商是联调或者测试类型，接口只能获取到20条数据
        $DB = new DB();
        $agent_info = $DB->row("SELECT account_type FROM lb_agent_user WHERE id = :id", ['id' => $agent_id]);
        if($agent_info['account_type'] != 1)
        {
            $take = (int)config('common.TEST_EXCEPTION_USER_COUNT');
        }

        $result = [];

        //根据最大ID获取数据
        if($deadline >= 0 && (empty($startDate) && empty($endDate)))
        {
            $deadline_utc = new \MongoDB\BSON\UTCDateTime(strtotime($deadline) * 1000);
            $findWhere = ['agent_id'=>(int)$agent_id];
            if($agent_info['account_type'] ==1 && $deadline > 0)
            {
                $findWhere['add_time'] = ['$gte'=>$deadline_utc];//时间条件
            }
            $result = Bootstrap::MongoDB()->selectCollection('live_game','exception_cash_log')->find(
                $findWhere,
                [
                    'projection' => [
                        'user_order_id' => 1, 
                        'user_name' => 1,
                        'agnet_name' => 1,
                        'hall_name' => 1,
                        'round_no' =>1,
                        'payout_win'   =>1,
                        'user_money'=>1,
                        'bet_time'   => 1,
                        'add_time'   => 1,
                        'action_user'   => 1,
                        'desc'   => 1,
                        '_id'   => 0
                    ],
                    'limit' => $take,
                    'sort'  => ['add_time'=>1]
                ]
            )->toArray();
        }

        //根据时间段获取数据
        if($startDate && $endDate)
        {
            $startDate_utc = strtotime($startDate) * 1000;
            $endDate_utc = strtotime($endDate) * 1000;
            $result = Bootstrap::MongoDB()->selectCollection('live_game','exception_cash_log')->find(
                [
                    'bet_time' =>['$gte'=> new \MongoDB\BSON\UTCDateTime($startDate_utc),'$lte'=>new \MongoDB\BSON\UTCDateTime($endDate_utc)],
                    'agent_id'  => (int)$agent_id
                ],
                [
                    'projection' => [
                        'user_order_id' => 1, 
                        'user_name' => 1,
                        'agnet_name' => 1,
                        'hall_name' => 1,
                        'round_no' =>1,
                        'payout_win'   =>1,
                        'user_money'=>1,
                        'bet_time'   => 1,
                        'add_time'   => 1,
                        'action_user'   => 1,
                        'desc'   => 1,
                        '_id'   => 0
                    ],
                    'limit' => $take,
                    'sort'  => ['bet_time'=>1]
                ]
            )->toArray();
        }

        return $result;
    }


    /**
     * 数据验证操作
     */
    private function makeValidate($byDate = false)
    {
        if(!$byDate)
        {
            $agent = Input::post('agent');
            $deadline = Input::post('deadline');
            $token = Input::post('token');
            $is_date = strtotime($deadline) ? strtotime($deadline) : 0;
            if(empty($agent)  || empty($token))
            {
                return false;
            }
            return true;
        }


    }

}