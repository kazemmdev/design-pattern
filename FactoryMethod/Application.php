<?php

namespace DesignPattern\FactoryMethods;

use Exception;

class Application
{
    private Logistics $logistics;

    /**
     * @throws Exception
     */
    public function __construct()
    {
        $this->initialize();
        $this->logistics->createDelivery();
    }

    /**
     * The application picks a creator's type depending on the
     * current configuration or environment settings.
     * @throws Exception
     */
    protected function initialize()
    {
        $config = $this->getConfigTransportation();

        if ($config['transport'] == 'sea') {
            $this->logistics = new SeaLogistics();
        } elseif ($config['transport'] == 'road') {
            $this->logistics = new RoadLogistics();
        } else {
            throw new Exception('Error! Unknown transport system.');
        }
    }

    protected function getConfigTransportation(): array
    {
        return [
            'transport' => 'road'
        ];
    }
}