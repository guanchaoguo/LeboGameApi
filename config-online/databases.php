<?php
/**
 * Created by PhpStorm.
 * User: liangxz@szljfkj.com
 * Date: 2017/3/14
 * Time: 14:17
 * 数据库配置文档
 */
return [

    'mysql' => [
        'database_type'    => 'mysql',
        'server'      => "10.200.66.76",
//        'port'      => 8066,
        'port'      => 3306,
//        'database_name'  => "le",
        'database_name'  => "lb_livegame",
        'username'  => "svr_livegame",
        'password'  => "yOikuPYpwUL5KjjKod01",
        'charset'   => "utf8",
    ],


    'master' => [
        'mysql'=>[
            'database_type'    => 'mysql',
            'server'      => "10.200.66.76",
            'port'      => 3306,
            'database_name'  => "lb_livegame",
            'username'  => "svr_livegame",
            'password'  => "yOikuPYpwUL5KjjKod01",
            'charset'   => "utf8",
        ]
    ],
    'slave' => [
        'mysql'=>[
            'database_type'    => 'mysql',
            'server'      => "10.100.205.103",
            'port'      => 3306,
            'database_name'  => "lb_livegame",
            'username'  => "svr_livegame",
            'password'  => "yOikuPYpwUL5KjjKod01",
            'charset'   => "utf8",
        ]
    ],


    'mongodb' => [
        'driver'   => 'mongodb',
        'host'     => 'mongodb://lg_web:IhIyMWmawP8WbvJAb6Fw@10.200.66.11:30000,10.200.66.12:30000,10.200.66.13:30000/live_game',
        'port'     => 30000,
        'database' => "live_game",
        'username' => "lg_web",
        'password' => "IhIyMWmawP8WbvJAb6Fw",
//        'replicaSet' => 'LiveGameRepl',
        'db' => 'live_game', // sets the authentication database required by mongo 3
    ],

    'redis' => [
         'default' =>  ['host'     => '10.200.66.72', 'port'     => 6379, 'password' => 'LiveGame2017',], //代理商名称和厅主白名单关系
         'monitor' =>  ['host'     => '10.200.66.73', 'port'     => 6379, 'password' => 'LiveGame2017',],// 用户增加充值、扣款后累计下注次数；
         'account' =>  ['host'     => '10.200.66.74', 'port'     => 6379, 'password' => 'LiveGame2017',],//已登录过用户数据
    ],
//    'redis' => [
//
//        'cluster' => env('REDIS_CLUSTER', false),
//
//        'default' => [
//            'host'     => env('REDIS_HOST', '127.0.0.1'),
//            'port'     => env('REDIS_PORT', 6379),
//            'database' => env('REDIS_DATABASE', 0),
//            'password' => env('REDIS_PASSWORD', null),
//        ],
//
//    ],
     'rabbitmq' => [
        'host'     => '10.200.66.95',//MQ服务器ip
        'port'     => '5672',//MQ服务器端口
        'name'     => 'gameApi',//MQ队列名称
        'username' => "lebo2017",//登录用户名
        'password' => 'yOikuPYpwUL5KjjKod01',//登录密码
        'vhost' => "leboGame",//虚拟机
        'channel' => "PhpApiLogExchange",//交换机
        'key' => "gameapi",//路由key
    ],

];