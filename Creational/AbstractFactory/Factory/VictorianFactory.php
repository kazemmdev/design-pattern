<?php

namespace DesignPattern\AbstractFactory\Factory;

use DesignPattern\AbstractFactory\Product\Chair;
use DesignPattern\AbstractFactory\Product\Sofa;
use DesignPattern\AbstractFactory\Product\VictorianChair;
use DesignPattern\AbstractFactory\Product\VictorianSofa;

class VictorianFactory implements FurnitureFactory
{

    public function createChair(): Chair
    {
        return new VictorianChair();
    }

    public function createSofa(): Sofa
    {
        return new VictorianSofa();
    }
}