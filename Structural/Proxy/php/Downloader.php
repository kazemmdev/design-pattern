<?php

namespace DesignPatterns\Structural\Proxy;

/**
 * The Subject interface. Both the real service and the proxy implement it, which
 * is what makes them interchangeable from the client's point of view.
 */
interface Downloader
{
    public function download(string $url): string;
}
