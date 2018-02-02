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

class UserCashController
{

    public function __construct()
    {
        IpLimit::init();//白名单过滤
        SecurityKeyMiddleware::handle();//SecurityKey过滤

    }
    /**
     * @SWG\Post(
     *   path="/orderList",
     *   tags={"api"},
     *   summary="获取注单信息",
     *   description="
     * 获取代理商每局游戏注单信息
     * PS:需要注意的是：第一次获取时时间参数为空，当获取数据后，调用方必须要进行数据落地；
     * 而后把数据最后一条时间作为参数请求，获取下一批次时间数据信息
     *
     * 接口信息返回说明：
     * code： 0 成功；
     * text ：'接口错误信息描述'；
     * result {
     *     data : [
     *        next_time：下一个批次的时间参数
     *        {
     *          game_round_id：游戏局号;
     *          total_bet_score：总下注金额;
     *          total_win_score：游戏结果(-100 or 100);
     *          valid_bet_score_total：有效投注金额;
     *          game_id：游戏ID;
     *          table_no：桌号;
     *          start_time：游戏时间;
     *          game_hall_id：游戏厅类型,0:旗舰厅，1贵宾厅，2：金臂厅， 3：至尊厅
     *          game_name: 游戏名称
     *          game_result: 游戏开牌数据
     *          }
     *       ]
     *      }
     *     ",
     *   operationId="index",
     *   @SWG\Parameter(
     *     in="formData",
     *     name="agent",
     *     type="string",
     *     description="代理商用户名",
     *     required=true,
     *     default="agent_test"
     *   ),
     *   @SWG\Parameter(
     *     in="formData",
     *     name="deadline",
     *     type="string",
     *     format="date",
     *     description="时间",
     *     required=true,
     *     default="2017-02-24 10:09:08"
     *   ),
     *   @SWG\Parameter(
     *     in="formData",
     *     name="token",
     *     type="string",
     *     description="SHA1('securityKey|deadline|agent')",
     *     required=true,
     *     default="46049a5df4b3b8395f6e3c0dc7d154315a2739b0"
     *   ),
     *   @SWG\Response(response="200", description="Success")
     * )
     */
    public function orderList()
    {
        $apiLog['start_time'] = time();
        $apiLog['user_name'] = Input::post('agent');
        $apiLog['postData'] = json_encode($_REQUEST);
        $apiLog['apiName'] = '获取注单信息';
        $apiLog['ip_info'] = get_real_ip();
        $apiLog['log_type'] = 'api';


        $agent = Input::post('agent');
        $deadline = Input::post('deadline');
        $token = Input::post('token');

        // 联调状态统计 只要有联调请求则统计数据
        $debugging = false;
        if(ApiStatistics($apiLog)) $debugging = true;


        //var_export(sha1(getSecurityKey($agent)."|".$deadline."|".$agent));die;
        //判断调用频率间隔时间
        if($feedError = $this->getListFeed($agent))
        {
            $apiLog['code'] = 9008;
            $apiLog['text'] = config('errorcode.9008');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

            return $feedError;
        }

        //数据验证错误
        if(!$this->makeValidate())
        {
            $apiLog['code'] = 9002;
            $apiLog['text'] = config('errorcode.9002');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

            return returnError();
        }

        //判断系统是否在维护当中
        if( $maintain = IpLimit::checkIsMaintain(Input::post('username'), Input::post('agent')) )
        {
            return $maintain;
        }

        //token验证，判断数据是否被篡改
        $param = [getSecurityKey($agent),$deadline,$agent];
        if(!checkEequest($token,$param))
        {
            $apiLog['code'] = 9002;
            $apiLog['text'] = config('errorcode.9002');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

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
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

            return returnJson([
                'code'  => 9004,
                'text'  => config('errorcode.9004'),
                'result'   => ''
            ]);
        }

        //验证通过进行数据获取操作
        $result = $this->getOrderList($agentInfo['id'],$deadline);

        //数据为空，返回空数据提示
        if(!$result)
        {
            $apiLog['code'] = 9009;
            $apiLog['text'] = config('errorcode.9009');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
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
            if(isset($val['start_time']) && is_object($val['start_time'])  ){
                $start_date = date("Y-m-d H:i:s",ceil($val['start_time']->__toString()/1000));
                $result[$key]['_id'] = $val['_id']->__toString();
                $result[$key]['start_time'] =  $start_date;
                if(($key+1) == count($result)) $res['next_time'] = $start_date;
            }
        }
        $res['data'] = $result;

        $apiLog['code'] = 0;
        $apiLog['text'] = config('errorcode.9999');
        $apiLog['result'] = json_encode($result);
        $apiLog['end_time'] = time();
        addApiLog($apiLog);//日志记录
        if($debugging)  ApiSucceds($apiLog);//联调账号联调成功次数统计

        //正常返回数据
        return returnJson([
            'code'      => 0,
            'text'      => config('errorcode.9999'),
            'result'    => $res
        ]);
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

    /**
     * @SWG\Post(
     *   path="/getDateList",
     *   tags={"api"},
     *   summary="获取注单信息",
     *   description="
     * 时间段获取代理商每局游戏注单信息
     * PS:需要注意的是：开始时间和结束时间跨度甭能超过一个星期
     *
     * 接口信息返回说明：
     * code： 0 成功；
     * text ：'接口错误信息描述'；
     * result {
     *     data : [
     *        next time：下一个批次的时间参数
     *        {
     *          game_round_id：游戏局号;
     *          total_bet_score：总下注金额;
     *          total_win_score：游戏结果(-100 or 100);
     *          valid_bet_score_total：有效投注金额;
     *          game_id：游戏ID;
     *          table_no：桌号;
     *          start_time：游戏时间;
     *          game_hall_id：游戏厅类型,0:旗舰厅，1贵宾厅，2：金臂厅， 3：至尊厅
     *          game_name: 游戏名称
     *          game_result: 游戏开牌数据
     *          }
     *       ]
     *      }
     *     ",
     *   operationId="index",
     *   @SWG\Parameter(
     *     in="formData",
     *     name="agent",
     *     type="string",
     *     description="代理商用户名",
     *     required=true,
     *     default="anchen2"
     *   ),
     *   @SWG\Parameter(
     *     in="formData",
     *     name="start_date",
     *     type="string",
     *     format="date",
     *     description="开始时间",
     *     required=true,
     *     default="2017-02-24 10:09:08"
     *   ),
     *   @SWG\Parameter(
     *     in="formData",
     *     name="end_date",
     *     type="string",
     *     format="date",
     *     description="结束时间",
     *     required=true,
     *     default="2017-02-24 10:09:08"
     *   ),
     *   @SWG\Parameter(
     *     in="formData",
     *     name="token",
     *     type="string",
     *     description="SHA1('securityKey|start_date|end_date|agent')",
     *     required=true,
     *     default="46049a5df4b3b8395f6e3c0dc7d154315a2739b0"
     *   ),
     *   @SWG\Response(response="200", description="Success")
     * )
     */
    public function getOrderListByDate()
    {
        $apiLog['start_time'] = time();
        $apiLog['user_name'] = Input::post('agent');
        $apiLog['postData'] = json_encode($_REQUEST);
        $apiLog['apiName'] = '时间段获取注单信息';
        $apiLog['ip_info'] = get_real_ip();
        $apiLog['log_type'] = 'api';

        $agent = Input::post('agent');
        $start_date = Input::post('start_date');
        $end_date = Input::post('end_date');
        $token = Input::post('token');

        // 联调状态统计 只要有联调请求则统计数据
        $debugging = false;
        if(ApiStatistics($apiLog)) $debugging = true;


        //var_export(sha1(getSecurityKey($agent)."|".$start_date."|".$end_date."|".$agent));die;
        //判断调用频率间隔时间
        if($feedError = $this->getListFeed($agent))
        {
            $apiLog['code'] = 9008;
            $apiLog['text'] = config('errorcode.9008');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

            return $feedError;
        }

        //数据验证错误
        if(!$this->makeDateValidate())
        {
            $apiLog['code'] = 9002;
            $apiLog['text'] = config('errorcode.9002');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
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
            $apiLog['end_time'] = time();
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
            $apiLog['end_time'] = time();
            addApiLog($apiLog);//日志记录

            return returnError();
        }

        //$agent = 'anchen2';
        //根据代理商获取代理商ID
//        $agentInfo = Bootstrap::DB()->get("lb_agent_user","*",['user_name'=>$agent]);
        $DB = new DB();
        $agentInfo = $DB->row("SELECT * FROM lb_agent_user WHERE user_name = :user_name and account_state  =:account_state", ['user_name'=>$agent ,'account_state'=> 1]);
        if(!$agentInfo)
        {
            $apiLog['code'] = 9004;
            $apiLog['text'] = config('errorcode.9004');
            $apiLog['result'] = '';
            $apiLog['end_time'] = time();
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
            $apiLog['end_time'] = time();
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
            if(isset($val['start_time']) && is_object($val['start_time'])  ){
                $start_date = date("Y-m-d H:i:s",ceil($val['start_time']->__toString()/1000));
                $result[$key]['start_time'] = $start_date;
                $result[$key]['_id'] = $val['_id']->__toString();
            }
        }

        $res['data'] = $result;

        $apiLog['code'] = 0;
        $apiLog['text'] = config('errorcode.9999');
        $apiLog['result'] = json_encode($result);
        $apiLog['end_time'] = time();
        addApiLog($apiLog);//日志记录
        if($debugging)  ApiSucceds($apiLog);//联调账号联调成功次数统计

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
            $take = (int)config('common.GET_ORDER_LIST_MAX');//获取单次最大获取的数据量条数，可配置

            //根据最大ID获取数据
            if($deadline >= 0 && (empty($startDate) && empty($endDate))) {
                $deadline_utc = new \MongoDB\BSON\UTCDateTime(strtotime($deadline) * 1000);
                $findWhere = ['agent_id'=>(int)$agent_id, 'is_mark'  => 1];
                if($deadline) {
                    $findWhere['end_time'] = ['$gte'=>$deadline_utc];//时间条件
                }
            }

            //根据时间段获取数据
            if($startDate && $endDate)
            {
                $startDate_utc = strtotime($startDate) * 1000;
                $endDate_utc = strtotime($endDate) * 1000;
                $findWhere = [
                    'end_time' =>['$gte'=> new \MongoDB\BSON\UTCDateTime($startDate_utc),'$lte'=>new \MongoDB\BSON\UTCDateTime($endDate_utc)],
                    'agent_id'  => (int)$agent_id,
                    'is_mark'  => 1,
                ];
            }
            $result = [];
            $field = [
                [ '$match' => $findWhere ],
                ['$project' => [
                        'round_no' => 1,
                        'game_id' => 1,
                        'game_hall_id' => 1,
                        'game_name' => 1,
                        'server_name' => 1,
                        'total_bet_score'   =>1,
                        'valid_bet_score_total'=>1,
                        'game_result'   => 1,
                        'total_win_score'=>1,
                        'start_time'    => '$end_time',
                        'user_name'=>1,
                        'is_mark'=>1,
                        'dwRound'=>1,
                        'remark'=>1,
                        '_id'   => 1
                    ],
                ],
                ['$limit' => $take],
                [ '$sort'  => ['end_time'=>1] ]

        ];
       $result = Bootstrap::MongoDB()->selectCollection('live_game','user_chart_info')->aggregate($field)->toArray();

       return $result;
    }

}