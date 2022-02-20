<?php

namespace DesignPatterns\Structural\Adapter\Tests;

use DesignPatterns\Structural\Adapter\Notification;
use DesignPatterns\Structural\Adapter\SlackApi;
use DesignPatterns\Structural\Adapter\SlackNotification;
use Orchestra\Testbench\TestCase;

class AdapterTest extends TestCase
{
    /** @test */
    public function is_slack_compatible_with_notification_class()
    {
        $slack = new SlackApi('k90mirzaei@gmail.com', 'XXXXXX');
        $notification = new SlackNotification($slack, '12345');

        $this->assertInstanceOf(Notification::class, $notification);

//        $notification->send('title', 'message');
    }
}