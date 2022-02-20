<?php

namespace DesignPatterns\Structural\Facade;

/**
 * The Facade provides a single method for downloading videos from YouTube. This
 * method hides all the complexity of the PHP network layer, YouTube API and the
 * video conversion library (FFmpeg).
 */
class YoutubeDownloader
{
    protected Youtube $youtube;

    protected \FFMpeg\FFMpeg $ffmpeg;

    public function __construct(string $youtubeApiKey)
    {
        $this->youtube = new Youtube($youtubeApiKey);
        $this->ffmpeg = \FFMpeg\FFMpeg::create();
    }

    /**
     * The Facade provides a simple method for downloading video and encoding it
     * to a target format (for the sake of simplicity, the real-world code is
     * commented-out).
     */
    public function download(string $url): void
    {
        $result = $this->youtube->fetchVideo($url);

        // Saving video file to a temporary file
        $this->youtube->saveAs($url, "video.mpg");

        // Processing source video
        $video = $this->ffmpeg->open('video.mpg');

        // Normalizing and resizing the video to smaller dimensions
        $video
            ->filters()
            ->resize(new \FFMpeg\Coordinate\Dimension(320, 240))
            ->synchronize();

        // Capturing preview image
        $video
            ->frame(\FFMpeg\Coordinate\TimeCode::fromSeconds(10))
            ->save($result . 'frame.jpg');

        // Saving video in target formats
        $video
            ->save(new \FFMpeg\Format\Video\X264(), $result . '.mp4')
            ->save(new \FFMpeg\Format\Video\WMV(), $result . '.wmv')
            ->save(new \FFMpeg\Format\Video\WebM(), $result . '.webm');
    }
}