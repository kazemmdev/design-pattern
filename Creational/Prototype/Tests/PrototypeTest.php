<?php

namespace DesignPatterns\Creational\Prototype\Tests;

use DesignPatterns\Creational\Prototype\Author;
use DesignPatterns\Creational\Prototype\Page;
use Orchestra\Testbench\TestCase;

class PrototypeTest extends TestCase
{
    /** @test */
    public function test_a_page_can_be_cloned_in_our_way()
    {
        $author = new Author("John Smith");
        $page = new Page("Tip of the day", "Keep calm and carry on.", $author);
        $page->addComment("Nice tip, thanks!");

        $draft = clone $page;

        $this->assertInstanceOf(Page::class, $draft);
    }
}