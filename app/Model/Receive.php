<?php

namespace App\Model;

class ReceiveController
{
    const MQHOST = '192.168.31.230'; //MQ服务器ip
    const PORT = '5672'; //MQ服务器端口
    const MQNAME = 'lebo'; //MQ队列名称
    const MQUSER = 'lebo2017'; //登录用户名
    const MQPWD = 'ljf12345'; //登录密码
    const MQVHOST = 'test'; //虚拟机
    const MQCHANNEL = 'e_linvo'; //交换机
    const MQKEY = 'gameapi'; //路由key

    private $mqobject = null;
    private $mqconn = null;
    private $mqqueue = null;

    public function __construct($mqhost = null, $mqprot = null, $mqname = null, $mquser = null, $mqpwd = null, $mqvhost = null)
    {
        $host = $mqhost ? $mqhost : self::MQHOST;
        $prot = $mqprot ? $mqprot : self::PORT;
        $name = $mqname ? $mqname : self::MQNAME;
        $user = $mquser ? $mquser : self::MQUSER;
        $pwd = $mqpwd ? $mqpwd : self::MQPWD;
        $vhost = $mqvhost ? $mqvhost : self::MQVHOST;

        self::connect($host, $prot, $name, $user, $pwd, $vhost);
    }

    //链接消息队列服务器
    private function connect($host, $prot, $name, $user, $pwd, $vhost)
    {
        $connect_list = ['host'=>$host, 'prot'=>$prot, 'login'=>$user, 'password'=>$pwd,'vhost'=>$vhost];
        $conn = new \AMQPConnection($connect_list);
        if (!$conn->connect()) {
            die("Cannot connect to the broker \n ");
        }
        $channel = new \AMQPChannel($conn);

        //创建交换机
        $e_name = self::MQCHANNEL; //交换机名
        $ex = new \AMQPExchange($channel);
        $ex->setName($e_name);
        $ex->setType(AMQP_EX_TYPE_DIRECT); //direct类型
        $ex->setFlags(AMQP_DURABLE); //持久化

        //创建队列
        $queue = new \AMQPQueue($channel);
        $queue->setName(self::MQNAME);
        $queue->setFlags(AMQP_DURABLE); //持久化
        //绑定交换机与队列，并指定路由键
        echo 'Queue Bind: '.$queue->bind($e_name, self::MQKEY)."\n";

        $this->mqobject = $ex;
        $this->mqconn = $conn;
        $this->mqqueue = $queue;
    }


    //进行消息拉取消费操作
    public function getMsg()
    {
        $ex = $this->mqobject ? $this->mqobject : $this->connect();
        //阻塞模式接收消息,只有消费端成功执行完成任务后，告诉MQ可以释放该消息
        echo "Message:\n";
        $this->mqqueue->consume(['App\Model\ReceiveController','processMessage'], AMQP_AUTOACK); //自动ACK应答

        $this->mqconn->disconnect();//断开链接
    }

    //拉取消息回调参数
    public static function processMessage($envelope, $queue)
    {
        $msg = $envelope->getBody();//获取消息队列消息
        echo $msg."\n"; //处理消息
    }
}

