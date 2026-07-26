<?php

namespace DesignPatterns\Structural\Proxy\Tests;

use DesignPatterns\Structural\Proxy\CachingDownloader;
use DesignPatterns\Structural\Proxy\SimpleDownloader;
use Orchestra\Testbench\TestCase;

class ProxyTest extends TestCase
{
    /** @test */
    public function proxy_returns_the_same_result_as_the_real_service()
    {
        $real = new SimpleDownloader();
        $proxy = new CachingDownloader(new SimpleDownloader());

        $this->assertEquals(
            $real->download('http://example.com/video.mp4'),
            $proxy->download('http://example.com/video.mp4')
        );
    }

    /** @test */
    public function repeated_downloads_hit_the_network_only_once()
    {
        $real = new SimpleDownloader();
        $proxy = new CachingDownloader($real);

        $proxy->download('http://example.com/video.mp4');
        $proxy->download('http://example.com/video.mp4');
        $proxy->download('http://example.com/video.mp4');

        $this->assertEquals(1, $real->downloads());
        $this->assertEquals(2, $proxy->hits());
    }

    /** @test */
    public function distinct_urls_are_cached_separately()
    {
        $real = new SimpleDownloader();
        $proxy = new CachingDownloader($real);

        $first = $proxy->download('http://example.com/a.mp4');
        $second = $proxy->download('http://example.com/b.mp4');

        $this->assertNotEquals($first, $second);
        $this->assertEquals(2, $real->downloads());
        $this->assertEquals(0, $proxy->hits());
    }

    /** @test */
    public function flushing_the_cache_forces_a_fresh_download()
    {
        $real = new SimpleDownloader();
        $proxy = new CachingDownloader($real);

        $proxy->download('http://example.com/video.mp4');
        $this->assertTrue($proxy->isCached('http://example.com/video.mp4'));

        $proxy->flush();
        $this->assertFalse($proxy->isCached('http://example.com/video.mp4'));

        $proxy->download('http://example.com/video.mp4');
        $this->assertEquals(2, $real->downloads());
    }
}
