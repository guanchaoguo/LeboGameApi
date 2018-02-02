<?php
/**
 * Created by PhpStorm.
 * User: liangxz@szljfkj.com
 * Date: 2017/3/20
 * Time: 16:08
 */
namespace App\controllers;

use bootstrap\Bootstrap;
use bootstrap\Input;
use lib\db\DB;

class SecurityKeyMiddleware
{
    public static function handle()
    {
        $agent = Input::post("agent");

        //  获取缓存数据中的数据是否命中名单信息
        $agentKeyName = 'agentWhitelist';
        if (!$ips = Bootstrap::Redis()->hGet($agentKeyName,$agent)) {
            if (!$agentInfo = getHallByAgent($agent)) {
                return returnJson([
                    'code' => 9004,//代理所属的厅主不存在
                    'text' => config('errorcode.9004'),
                    'result' => '',
                ]);
            }

            // 未命中缓存获取数据库白名单信息
            $whites = (new DB)->row("SELECT * FROM white_list WHERE agent_id = :agent_id AND state = :state", ["agent_id" => $agentInfo['parent_id'], 'state' => 1]);
            $ips = json_encode($whites);
        }

        //判断是否已经过期
        $result =json_decode($ips, true);
        if(strtotime($result['seckey_exp_date']) <= time()) {
            return returnJson([
                'code' => 9007,//SecurityKey已经过期
                'text' => config('errorcode.9007'),
                'result' => ''
            ]);
        }
    }

}