<?php

namespace DesignPatterns\Creational\AbstractFactory\Template\Page;

/**
 * This is another Abstract Product type, which describes whole page templates.
 */
interface PageTemplate
{
    public function getTemplateString(): string;
}