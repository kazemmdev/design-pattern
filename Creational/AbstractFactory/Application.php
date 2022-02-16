<?php

namespace DesignPattern\AbstractFactory;

use DesignPattern\AbstractFactory\Factory\FurnitureFactory;

class Application
{
    private FurnitureFactory $factory;

    public function __construct(FurnitureFactory $factory)
    {
        $this->factory = $factory;
    }


}