<?php

namespace App\Model;

use bootstrap\Bootstrap;

/**
 * Created by PhpStorm.
 * User: liangxz@szljfkj.com
 * Date: 2017/5/24
 * Time: 9:34
 *  MQ 消息生产者
 *  用于向消息队列 rabbitmq 发送消息
 */
class MqProducer
{
    private $mqobject = null;
    private $dsn = null;

    /*
     * 加载消息队列的参数
     */
    public function __construct($mqhost = null, $mqprot = null, $mqname = null, $mquser = null, $mqpwd = null, $mqvhost = null)
    {
        
        $this->dsn = Bootstrap::AMQPConfig();
        $host = $mqhost ? $mqhost :$this->dsn['host'];
        $prot = $mqprot ? $mqprot : $this->dsn['port'];
        $name = $mqname ? $mqname : $this->dsn['name'];
        $user = $mquser ? $mquser :$this->dsn['username'];
        $pwd = $mqpwd ? $mqpwd : $this->dsn['password'];
        $vhost = $mqvhost ? $mqvhost : $this->dsn['vhost'];

        self::connect($host, $prot, $name, $user, $pwd, $vhost);
    }

    /**
     * 连接消息队列服务器
     * @param $host  MQ服务器ip
     * @param $prot  MQ服务器端口
     * @param $name  MQ队列名称
     * @param $user  登录用户名
     * @param $pwd   登录密码
     * @param $vhost 虚拟机
     */
    private  function connect($host, $prot, $name, $user, $pwd, $vhost)
    {
        $connect_list = ['host'=>$host, 'prot'=>$prot, 'login'=>$user, 'password'=>$pwd,'vhost'=>$vhost];

        //创建消息队列连接
        $conn = new \AMQPConnection($connect_list);
        if (!$conn->connect()) {
            die("Cannot connect to the broker \n ");
        }

        //创建通信通道
        $channel = new \AMQPChannel($conn);

        //创建交换机
        $e_name = $this->dsn['channel']; //交换机名
        $ex = new \AMQPExchange($channel);
        $ex->setName($e_name);
        $ex->setType(AMQP_EX_TYPE_DIRECT); //direct类型
        $ex->setFlags(AMQP_DURABLE); //持久化

        //创建队列
        $queue = new \AMQPQueue($channel);
        $queue->setName( $this->dsn['name']);
        $queue->setFlags(AMQP_DURABLE); //持久化

        $this->mqobject = $ex;
    }

    /**
     * 进行消息推送操作
     * @param $data 向消息队列的发送的数据
     * @return mixed boolean 处理结果
     */
    public function publishMsg($data)
    {
        $ex = $this->mqobject ? $this->mqobject : $this->connect();
        return $ex->publish($data, $this->dsn['key']); //发送到指定的路由
    }
}
