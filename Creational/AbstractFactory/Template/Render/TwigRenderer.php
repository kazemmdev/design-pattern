<?php

namespace DesignPatterns\Creational\AbstractFactory\Template\Render;

/**
 * The renderer for Twig templates.
 */
class TwigRenderer implements TemplateRenderer
{
    public function render(string $templateString, array $arguments = []): string
    {
        $loader = new \Twig\Loader\ArrayLoader([
            'index.html' => $templateString,
        ]);

        $twig = new \Twig\Environment($loader);

        return $twig->render('index.html', $arguments);
    }
}