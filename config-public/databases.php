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
        'server'      => "10.200.124.23",
        'port'      => 3306,
        'database_name'  => "lb_livegame_test",
        'username'  => "Mysqladmin",
        'password'  => "Mysql1707!",
        'charset'   => "utf8",
        'option'    => [PDO::ATTR_PERSISTENT => true],//长连接方式进行
    ],


    'mongodb' => [
        'driver'   => 'mongodb',
        'host'     => 'mongodb://hhq163:bx123456@10.200.124.23/live_game',
    ],

    'redis' => [
        'default'=> ['host'     => '10.200.124.21', 'port'     => 6379, 'password' => 'bx123456',],//代理商名称和厅主白名单关系
        'monitor'=> ['host'     => '10.200.124.21', 'port'     => 6380, 'password' => 'bx123456',],// 用户增加充值、扣款后累计下注次数；
        'account'=> ['host'     => '10.200.124.21', 'port'     => 6381, 'password' => 'bx123456',],//已登录过用户数据
    ],

     'rabbitmq' => [
        'host'     => '10.200.124.21',//MQ服务器ip
        'port'     => '5672',//MQ服务器端口
        'name'     => 'gameApi',//MQ队列名称
        'username' => "lebo2017",//登录用户名
        'password' => 'ljf12345',//登录密码
        'vhost' => "test",//虚拟机
        'channel' => "PhpApiLogExchange",//交换机
        'key' => "gameapi",//路由key
    ],
];
