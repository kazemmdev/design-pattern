<?php

namespace DesignPatterns\Structural\Decorator\Tests;

use DesignPatterns\Structural\Decorator\DangerousHTMLTagsFilter;
use DesignPatterns\Structural\Decorator\MarkdownFormat;
use DesignPatterns\Structural\Decorator\PlainTextFilter;
use DesignPatterns\Structural\Decorator\TextInput;
use Orchestra\Testbench\TestCase;

class DecoratorTest extends TestCase
{
    protected string $dangerousComment;

    /** @test */
    public function check_native_text_format()
    {
        $native = new TextInput();

        $this->assertEquals($this->dangerousComment, $native->formatText($this->dangerousComment));
    }

    /** @test */
    public function check_plain_text_format()
    {
        $plain = new PlainTextFilter(new TextInput());

        $this->assertStringNotContainsString('<script>', $plain->formatText($this->dangerousComment));
    }

    /** @test */
    public function check_dangerous_html_format()
    {
        $plain = new DangerousHTMLTagsFilter(new TextInput());

        $this->assertStringNotContainsString('<script>', $plain->formatText($this->dangerousComment));
    }

    /** @test */
    public function check_markdown_format()
    {
        $plain = new DangerousHTMLTagsFilter(new MarkdownFormat(new TextInput()));
        $eval = $plain->formatText($this->dangerousComment);

        $this->assertStringNotContainsString('<script>', $eval);
        $this->assertStringContainsString('<h1>Welcome</h1>', $eval);
        $this->assertStringContainsString('<strong>gorgeous</strong>', $eval);
    }

    protected function setUp(): void
    {
        parent::setUp();
        $this->dangerousComment = <<<HERE
# Welcome

This is my first post on this **gorgeous** forum.

<script src="http://www.iwillhackyou.com/script.js">
  performXSSAttack();
</script>
HERE;
    }
}