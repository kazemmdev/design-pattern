<?php

namespace DesignPatterns\Creational\FactoryMethod\Tests;

use DesignPatterns\Creational\FactoryMethod\FacebookPoster;
use Orchestra\Testbench\TestCase;

class FactoryMethodTest extends TestCase
{
    /** @test */
    public function facebook_connector_test()
    {
        $x = new FacebookPoster('k90mirzaei@gmail.com', '123456789');

        $x->post('This is a sample post');

        $this->assertTrue(true);
    }
}