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
//        'port'      => 8066,
        'port'      => 3306,
//        'database_name'  => "le",
        'database_name'  => "lb_livegame",
        'username'  => "root",
        'password'  => "123456",
        'charset'   => "utf8",
        'option'    => [PDO::ATTR_PERSISTENT => true],//长连接方式进行
    ],


//    'master' => [
//        'mysql'=>[
//            'database_type'    => 'mysql',
//            'server'      => "10.200.66.76",
//            'port'      => 3306,
//            'database_name'  => "lb_livegame",
//            'username'  => "svr_livegame",
//            'password'  => "yOikuPYpwUL5KjjKod01",
//            'charset'   => "utf8",
//        ]
//    ],
//    'slave' => [
//        'mysql'=>[
//            'database_type'    => 'mysql',
//            'server'      => "10.100.205.103",
//            'port'      => 3306,
//            'database_name'  => "lb_livegame",
//            'username'  => "svr_livegame",
//            'password'  => "yOikuPYpwUL5KjjKod01",
//            'charset'   => "utf8",
//        ]
//    ],


    'mongodb' => [
        'driver'   => 'mongodb',
        'host'     => '192.168.31.231',
        'port'     => 27017,
        'database' => "live_game",
        'username' => "hhq163",
        'password' => "bx123456",
        'replicaSet' => 'LiveGameRepl',
        'db' => 'live_game', // sets the authentication database required by mongo 3
    ],

    'redis' => [
        'host'     => '192.168.31.230',
        'port'     => 6379,
        'password' => 'bx123456',
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

];