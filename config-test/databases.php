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
        'server'      => "192.168.31.231",
        'port'      => 3306,
        'database_name'  => "lb_livegame_test232",
        'username'  => "root",
        'password'  => "123456",
        'charset'   => "utf8",
        'option'    => [PDO::ATTR_PERSISTENT => true],//长连接方式进行
    ],


    'mongodb' => [
        'driver'   => 'mongodb', 
        'host'     => "mongodb://hhq163:bx123456@192.168.31.231:27017/live_game_test232",
        'port'     => 27017,
        'database' => "live_game_test232",
        'username' => "hhq163",
        'password' => "bx123456",
        'db' => 'live_game', // sets the authentication database required by mongo 3
    ],

    'redis' => [
          'default'=> ['host'     => '192.168.31.232', 'port'     => 6379, 'password' => 'bx123456',], //代理商名称和厅主白名单关系
          'monitor'=> ['host'     => '192.168.31.232', 'port'     => 6380, 'password' => 'bx123456',], // 用户增加充值、扣款后累计下注次数；
          'account'=> ['host'     => '192.168.31.232', 'port'     => 6381, 'password' => 'bx123456',], //已登录过用户数据
    ],
     'rabbitmq' => [
        'host'     => '192.168.31.230',//MQ服务器ip
        'port'     => '5672',//MQ服务器端口
        'name'     => 'gameApi',//MQ队列名称
        'username' => "lebo2017",//登录用户名
        'password' => 'ljf12345',//登录密码
        'vhost' => "test",//虚拟机
        'channel' => "PhpApiLogExchange",//交换机
        'key' => "gameapi",//路由key
    ],

];
