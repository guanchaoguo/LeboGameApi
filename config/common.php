<?php
/**
 * Created by PhpStorm.
 * User: liangxz@szljfkj.com
 * Date: 2017/3/15
 * Time: 15:59
 * 框架业务公共配置信息
 * PS:目前只支持一维数组配置格式
 */
return [
    "GAME_API_SUF"  => "leboapi",//#游戏api后缀
    "AGENT_IPS"     => ":ips",//代理商ip列表标识
    "AGENT_ORDER_LIST"  => '_get_order_list',//获取代理商用户注单信息后缀
    "GET_ORDER_LIST_SPEED"  => 1,//限定获取注单信息接口的时间间隔：秒
    "GET_ORDER_LIST_MAX"    => 1000,//限定每次获取注单数据最大记录的条数
    "MAX_TIME_SPAN"     => 7,//限定获取注单信息的最大时间跨度：天
    "KEY_MAX_VALID_TIME"     => 30,//SecurityKey最大有效时间:天
    "SECURITY_KEY_ENCRYPT"  => 'sha1',//SecurityKey加密方式，默认为：sha1,可选项为sha256
    'GAME_HOST' => 'http://h5.lggame.co',//游戏客户端地址
    'GAME_HOST_PC' => 'https://pc.lggame.co/game.php',//游戏PC端地址
    'TEST_USER_COUNT' => 5, //测试联调账号只能获取5条注单数据
];

