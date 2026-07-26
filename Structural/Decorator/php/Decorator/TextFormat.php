<?php

namespace DesignPatterns\Structural\Decorator;

class TextFormat implements InputFormat
{
    private InputFormat $format;

    public function __construct(InputFormat $format)
    {
        $this->format = $format;
    }

    public function formatText(string $text): string
    {
        return $this->format->formatText($text);
    }
}