<?php

namespace DesignPatterns\Structural\Proxy;

/**
 * The Real Subject. This is the class that does the actual (expensive) work.
 * Downloading the same file twice costs twice the bandwidth and twice the wait.
 */
class SimpleDownloader implements Downloader
{
    private int $downloads = 0;

    public function download(string $url): string
    {
        // A real implementation would issue an HTTP request here. The counter
        // stands in for that cost so the proxy's effect can be observed.
        $this->downloads++;

        return "content of $url";
    }

    /**
     * How many times the network was actually hit.
     */
    public function downloads(): int
    {
        return $this->downloads;
    }
}
