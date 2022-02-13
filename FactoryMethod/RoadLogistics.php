<?php

namespace DesignPattern\FactoryMethods;

class RoadLogistics extends Logistics
{
    public function Create(): Transport
    {
        return new Truck();
    }
}