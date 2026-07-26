<?php

namespace DesignPatterns\Creational\AbstractFactory\Factory;

use DesignPatterns\Creational\AbstractFactory\Template\Page\PageTemplate;
use DesignPatterns\Creational\AbstractFactory\Template\Render\TemplateRenderer;
use DesignPatterns\Creational\AbstractFactory\Template\Title\TitleTemplate;

/**
 * The Abstract Factory interface declares creation methods for each distinct
 * product type.
 */
interface TemplateFactory
{
    public function createTitleTemplate(): TitleTemplate;

    public function createPageTemplate(): PageTemplate;

    public function getRenderer(): TemplateRenderer;
}