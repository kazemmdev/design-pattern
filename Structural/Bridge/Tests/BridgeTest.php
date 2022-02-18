<?php

namespace DesignPatterns\Structural\Bridge\Tests;

use DesignPatterns\Structural\Bridge\src\Abstraction\Product;
use DesignPatterns\Structural\Bridge\src\Abstraction\ProductPage;
use DesignPatterns\Structural\Bridge\src\Abstraction\SimplePage;
use DesignPatterns\Structural\Bridge\src\Implementation\HTMLRenderer;
use DesignPatterns\Structural\Bridge\src\Implementation\JsonRenderer;
use Orchestra\Testbench\TestCase;

class BridgeTest extends TestCase
{
    /** @test */
    public function is_simple_page_has_title_in_html_renderer()
    {
        $simple = new SimplePage(new HTMLRenderer(), 'title', 'content');
        $html = $simple->view();

        $this->assertMatchesRegularExpression('/<h1>title<\/h1>/', $html);
    }

    /** @test */
    public function is_simple_page_has_title_in_json_renderer()
    {
        $simple = new SimplePage(new JsonRenderer(), 'title', 'content');
        $json = json_decode($simple->view());

        $this->assertEquals('title', $json->title);
        $this->assertEquals('content', $json->text);
    }

    /** @test */
    public function can_simple_page_change_renderer()
    {
        $simple = new SimplePage(new HTMLRenderer(), 'title', 'content');
        $simple->changeRenderer(new JsonRenderer());

        $json = json_decode($simple->view());

        $this->assertEquals('title', $json->title);
        $this->assertEquals('content', $json->text);
    }

    /** @test */
    public function is_product_page_has_title_in_html_renderer()
    {
        $product = new Product("123", "Star Wars, episode1",
            "A long time ago in a galaxy far, far away...",
            "/images/star-wars.jpeg", 39.95);

        $productPage = new ProductPage(new HTMLRenderer(), $product);

        $html = $productPage->view();

        $this->assertMatchesRegularExpression('/<h1>Star Wars, episode1<\/h1>/', $html);
    }

    /** @test */
    public function is_product_page_has_title_in_json_renderer()
    {
        $product = new Product("123", "Star Wars, episode1",
            "A long time ago in a galaxy far, far away...",
            "/images/star-wars.jpeg", 39.95);

        $productPage = new ProductPage(new JsonRenderer(), $product);

        $json = json_decode($productPage->view());

        $this->assertEquals('Star Wars, episode1', $json->title);
        $this->assertEquals('A long time ago in a galaxy far, far away...', $json->text);
    }

    /** @test */
    public function can_product_page_change_renderer()
    {
        $product = new Product("123", "Star Wars, episode1",
            "A long time ago in a galaxy far, far away...",
            "/images/star-wars.jpeg", 39.95);

        $productPage = new ProductPage(new JsonRenderer(), $product);
        $productPage->changeRenderer(new JsonRenderer());

        $json = json_decode($productPage->view());

        $this->assertEquals('Star Wars, episode1', $json->title);
        $this->assertEquals('A long time ago in a galaxy far, far away...', $json->text);
    }
}