<?php

namespace DesignPatterns\Creational\Builder;

class QueryDirector
{
    private SQLQueryBuilder $builder;

    public function __construct(SQLQueryBuilder $builder)
    {
        if (config('database.type') == 'mysql') {
            $this->builder = new MysqlQueryBuilder();
        } elseif (config('database.type') == 'postgres') {
            $this->builder = new PostgresQueryBuilder();
        } else {
            throw new \Exception("Database type isn't configure");
        }
    }

    public function getYoungUser(): string
    {
        return $this->builder->select("users", ["name", "email"])
            ->where("age", 20, "<")
            ->limit(10, 20)
            ->getSQL();
    }

    public function getYoungUserByBuilder(SQLQueryBuilder $builder): string
    {
        return $builder->select("users", ["name", "email"])
            ->where("age", 20, "<")
            ->limit(10, 20)
            ->getSQL();
    }
}