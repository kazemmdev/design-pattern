<?php

namespace DesignPatterns\Structural\Flyweight\Tests;

use DesignPatterns\Structural\Flyweight\CatDataBase;
use Orchestra\Testbench\TestCase;

class FlyweightTest extends TestCase
{
    /** @test */
    public function test_cat_variations_are_same()
    {
        $db = new CatDataBase();
        $db->addCat('name1', 'age1', 'owner1',
            'breed', 'image', 'color', 'texture', 'fur', 'size');
        $db->addCat('name2', 'age2', 'owner2',
            'breed', 'image', 'color', 'texture', 'fur', 'size');

        $cat1 = $db->findCat(['name' => 'name1']);
        $cat2 = $db->findCat(['name' => 'name2']);

        $this->assertSame($cat1->getVariation(), $cat2->getVariation());
    }
}