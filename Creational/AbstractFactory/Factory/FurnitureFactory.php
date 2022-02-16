<?php

namespace DesignPattern\AbstractFactory\Factory;

use DesignPattern\AbstractFactory\Product\Chair;
use DesignPattern\AbstractFactory\Product\Sofa;

// The abstract factory interface declares a set of methods that
// return different abstract products. These products are called a family.
// A family of products may have several variants,
// but the products of one variant are incompatible with the
// products of another variant.

interface FurnitureFactory
{
    public function createChair(): Chair;

    public function createSofa(): Sofa;
}