<?php

namespace DesignPatterns\Creational\Singleton\Tests;

use DesignPatterns\Creational\Singleton\Config;
use DesignPatterns\Creational\Singleton\Logger;
use Orchestra\Testbench\TestCase;

class SingletonTest extends TestCase
{
    /** @test */
    public function is_instances_of_logger_are_same()
    {
        $l1 = Logger::getInstance();
        $l2 = Logger::getInstance();

        $this->assertEquals($l1, $l2);
    }

    /** @test */
    public function is_config_save_data_correctly()
    {
        $config1 = Config::getInstance();
        $login = "test_login";
        $password = "test_password";
        $config1->setValue("login", $login);
        $config1->setValue("password", $password);

        $config2 = Config::getInstance();

        $this->assertEquals($login, $config2->getValue('login'));
        $this->assertEquals($password, $config2->getValue('password'));

//        dd(
//            Logger::log('Hi'),
//            file_get_contents('php://stdout')
//        );
    }

}