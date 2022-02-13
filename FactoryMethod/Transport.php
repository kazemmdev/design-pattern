<?php

namespace DesignPattern\FactoryMethods;

interface Transport
{
    // The product interface declares the operations that all
    // concrete products must implement.
    public function deliver();
}