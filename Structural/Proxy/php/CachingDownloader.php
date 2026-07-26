<?php

namespace DesignPatterns\Structural\Proxy;

/**
 * The Proxy. It implements the same interface as the Real Subject, so the client
 * cannot tell them apart, and it keeps a reference to the service it wraps.
 *
 * Here the added behaviour is caching: the first request for a URL is delegated
 * to the real downloader, every later request for the same URL is served from
 * memory without touching the network.
 */
class CachingDownloader implements Downloader
{
    private Downloader $downloader;

    /** @var array<string, string> */
    private array $cache = [];

    private int $hits = 0;

    public function __construct(Downloader $downloader)
    {
        $this->downloader = $downloader;
    }

    public function download(string $url): string
    {
        if (! array_key_exists($url, $this->cache)) {
            // Cache miss: do the real work, then remember it.
            $this->cache[$url] = $this->downloader->download($url);

            return $this->cache[$url];
        }

        $this->hits++;

        return $this->cache[$url];
    }

    /**
     * How many requests were served from the cache instead of the network.
     */
    public function hits(): int
    {
        return $this->hits;
    }

    public function isCached(string $url): bool
    {
        return array_key_exists($url, $this->cache);
    }

    /**
     * Drop the cached responses. A real proxy would also do this on a TTL.
     */
    public function flush(): void
    {
        $this->cache = [];
    }
}
