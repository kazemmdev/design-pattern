<?php

namespace DesignPatterns\Creational\FactoryMethod\Tests;

use DesignPatterns\Creational\FactoryMethod\Factory\FacebookPoster;
use DesignPatterns\Creational\FactoryMethod\Factory\LinkedInPoster;
use Orchestra\Testbench\TestCase;

class FactoryMethodTest extends TestCase
{
    /** @test */
    public function facebook_connector_test()
    {
        $x = new FacebookPoster('k90mirzaei@gmail.com', '123456789');
        $response = $x->post('This is a sample post');

        $this->assertEquals(200, $response->getStatusCode());
    }

    /** @test */
    public function linkedin_connector_test()
    {
        $x = new LinkedInPoster('k90mirzaei@gmail.com', '123456789');
        $response = $x->post('This is a sample post');

        $this->assertEquals(200, $response->getStatusCode());
    }
}