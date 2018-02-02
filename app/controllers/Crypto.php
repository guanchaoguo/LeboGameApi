<?php
/**
 * 加密解密类.
 * User: chensongjian
 * Date: 2017/4/21
 * Time: 14:45
 */

namespace App\controllers;


class Crypto
{
    const ralphabet = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890_';
    const alphabet = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890_';

    /**
     * @param string $strtoencrypt
     * @return string
     */
    public static function encrypt (string $strtoencrypt) : string
    {
        $password = "kinge383e";
        for( $i=0; $i<strlen($password); $i++ ) {
            $cur_pswd_ltr = substr($password,$i,1);
            $pos_alpha_ary[] = substr(strstr(self::alphabet,$cur_pswd_ltr),0,strlen(self::ralphabet));
        }

        $i=0;
        $n = 0;
        $nn = strlen($password);
        $c = strlen($strtoencrypt);
        $encrypted_string = '';

        while($i<$c) {
            $encrypted_string .= substr($pos_alpha_ary[$n],strpos(self::ralphabet,substr($strtoencrypt,$i,1)),1);
            $n++;
            if($n==$nn) $n = 0;
            $i++;
        }

        return $encrypted_string;

    }

    public static function decrypt ($strtodecrypt)
    {
        $password = "kinge383e";
        for( $i=0; $i<strlen($password); $i++ ) {
            $cur_pswd_ltr = substr($password,$i,1);
            $pos_alpha_ary[] = substr(strstr(self::alphabet,$cur_pswd_ltr),0,strlen(self::ralphabet));
        }

        $i=0;
        $n = 0;
        $nn = strlen($password);
        $c = strlen($strtodecrypt);
        $decrypted_string = '';

        while($i<$c) {
            $decrypted_string .= substr(self::ralphabet,strpos($pos_alpha_ary[$n],substr($strtodecrypt,$i,1)),1);
            $n++;
            if($n==$nn) $n = 0;
            $i++;
        }

        return $decrypted_string;

    }
}