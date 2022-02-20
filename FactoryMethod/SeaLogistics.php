<?php

namespace DesignPattern\FactoryMethods;

class SeaLogistics extends Logistics
{
    public function Create(): Transport
    {
        return new Ship();
    }
}