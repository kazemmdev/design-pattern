<?php

namespace DesignPatterns\Structural\Decorator;

interface InputFormat
{
    public function formatText(string $text): string;
}