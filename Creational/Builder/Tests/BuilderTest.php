<?php

namespace DesignPatterns\Creational\FactoryMethod\Tests;

use DesignPatterns\Creational\Builder\MysqlQueryBuilder;
use DesignPatterns\Creational\Builder\PostgresQueryBuilder;
use DesignPatterns\Creational\Builder\QueryDirector;
use Orchestra\Testbench\TestCase;

class BuilderTest extends TestCase
{
    protected string $query;

    protected function setUp(): void
    {
        parent::setUp();

        config()->set('database.type', 'mysql');
        $this->query = "SELECT name, email FROM users WHERE age < '20' LIMIT 10, 20;";
    }

    /** @test */
    public function test_query_mysql_builder()
    {
        $query = new QueryDirector(new MysqlQueryBuilder());

        $this->assertEquals($this->query, $query->getYoungUser());
    }

    /** @test */
    public function test_query_postgres_builder()
    {
        $query = new QueryDirector(new PostgresQueryBuilder());

        $this->assertEquals($this->query, $query->getYoungUser());
    }
}