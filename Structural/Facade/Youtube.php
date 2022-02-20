<?php

namespace DesignPatterns\Structural\Facade;

class Youtube
{
    public function fetchVideo(string $url): string
    {
        return 'Saved';
    }

    public function saveAs(string $path): void
    {
        // save video in storage
    }
}