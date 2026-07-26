<?php

namespace DesignPatterns\Creational\AbstractFactory\Factory;

use DesignPatterns\Creational\AbstractFactory\Template\Page\PageTemplate;
use DesignPatterns\Creational\AbstractFactory\Template\Page\PHPTemplatePageTemplate;
use DesignPatterns\Creational\AbstractFactory\Template\Render\PHPTemplateRenderer;
use DesignPatterns\Creational\AbstractFactory\Template\Render\TemplateRenderer;
use DesignPatterns\Creational\AbstractFactory\Template\Title\PHPTemplateTitleTemplate;
use DesignPatterns\Creational\AbstractFactory\Template\Title\TitleTemplate;

/**
 * And this Concrete Factory creates PHPTemplate templates.
 */
class PHPTemplateFactory implements TemplateFactory
{
    public function createTitleTemplate(): TitleTemplate
    {
        return new PHPTemplateTitleTemplate();
    }

    public function createPageTemplate(): PageTemplate
    {
        return new PHPTemplatePageTemplate($this->createTitleTemplate());
    }

    public function getRenderer(): TemplateRenderer
    {
        return new PHPTemplateRenderer();
    }
}