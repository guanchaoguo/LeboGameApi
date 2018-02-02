<?php
/**
 * Created by PhpStorm.
 * User: liangxz@szljfkj.com
 * Date: 2017/3/22
 * Time: 14:46
 */

namespace App\controllers;


use bootstrap\Bootstrap;
use lib\db\DB;

class Mycat
{
    public function mycat()
    {
////        var_export(hash('sha1','981ef916a0dd7f1802f6ff5867abf4d6f3eb1539|0|agent_test'));die;
//        $dsn = "mysql:dbname=lb_livegame;host=192.168.31.232;port=3306;charset=utf8";
//        $conn = new \PDO($dsn,"root","123456");
//        $sql = "select * from menus";
//        foreach ($conn->query($sql) as $row) {
//            print_r($row); //你可以用 echo($GLOBAL); 来看到这些值
//        }
//        die;
//        $res = Bootstrap::DB()->select('menus','*');
//        var_export($res);
        $db = new DB();
        $res = $db->query("select * from menus");
        foreach ($res as $val)
        {
            print_r($val);
        }
    }
}