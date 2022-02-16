<?php

namespace DesignPatterns\Creational\AbstractMethod\Tests;

use DesignPatterns\Creational\AbstractFactory\Factory\PHPTemplateFactory;
use DesignPatterns\Creational\AbstractFactory\Factory\TwigTemplateFactory;
use DesignPatterns\Creational\AbstractMethod\Page;
use Orchestra\Testbench\TestCase;

class AbstractFactoryTest extends TestCase
{
    protected Page $page;
    protected string $htmlPage;

    /** @test */
    public function php_page_render_test()
    {
        $html = $this->page->render(new PHPTemplateFactory());
        $this->assertEquals($this->htmlPage, $html);
    }

    /** @test */
    public function twig_page_render_test()
    {
        $html = $this->page->render(new TwigTemplateFactory());
        $this->assertEquals($this->htmlPage, $html);
    }

    protected function setUp(): void
    {
        parent::setUp();

        $this->page = new Page('Sample page', 'This is the body.');
        $this->htmlPage = '<div class="page">
    <h1>Sample page</h1>
    <article class="content">This is the body.</article>
</div>';
    }
}