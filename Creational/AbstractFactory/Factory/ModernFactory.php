<?php

namespace DesignPattern\AbstractFactory\Factory;

use DesignPattern\AbstractFactory\Product\Chair;
use DesignPattern\AbstractFactory\Product\ModernChair;
use DesignPattern\AbstractFactory\Product\ModernSofa;
use DesignPattern\AbstractFactory\Product\Sofa;

class ModernFactory implements FurnitureFactory
{

    public function createChair(): Chair
    {
        return new ModernChair();
    }

    public function createSofa(): Sofa
    {
        return new ModernSofa();
    }
}