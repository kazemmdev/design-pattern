<?php

namespace DesignPatterns\Creational\AbstractFactory\Factory;

use DesignPatterns\Creational\AbstractFactory\Template\Page\PageTemplate;
use DesignPatterns\Creational\AbstractFactory\Template\Page\TwigPageTemplate;
use DesignPatterns\Creational\AbstractFactory\Template\Render\TemplateRenderer;
use DesignPatterns\Creational\AbstractFactory\Template\Render\TwigRenderer;
use DesignPatterns\Creational\AbstractFactory\Template\Title\TitleTemplate;
use DesignPatterns\Creational\AbstractFactory\Template\Title\TwigTitleTemplate;

/**
 * Each Concrete Factory corresponds to a specific variant (or family) of
 * products.
 *
 * This Concrete Factory creates Twig templates.
 */
class TwigTemplateFactory implements TemplateFactory
{
    public function createTitleTemplate(): TitleTemplate
    {
        return new TwigTitleTemplate();
    }

    public function createPageTemplate(): PageTemplate
    {
        return new TwigPageTemplate($this->createTitleTemplate());
    }

    public function getRenderer(): TemplateRenderer
    {
        return new TwigRenderer();
    }
}
