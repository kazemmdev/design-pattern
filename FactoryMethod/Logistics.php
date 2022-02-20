<?php

namespace DesignPattern\FactoryMethods;

abstract class Logistics
{
    // The creator class declares the factory method that must
    // return an object of a product class. The creator's subclasses
    // usually provide the implementation of this method.
    abstract public function Create(): Transport;

    public function createDelivery()
    {
        // Call the factory method to create a product object.
        $transport = $this->Create();

        // Now use the product.
        $transport->deliver();
    }
}