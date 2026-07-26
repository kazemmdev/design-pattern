<?php

namespace DesignPatterns\Structural\Adapter;

/**
 * The Adaptee is some useful class, incompatible with the Target interface. You
 * can't just go in and change the code of the class to follow the Target
 * interface, since the code might be provided by a 3rd-party library.
 */
class SlackApi
{
    private string $login;
    private string $apiKey;

    public function __construct(string $login, string $apiKey)
    {
        $this->login = $login;
        $this->apiKey = $apiKey;
    }

    public function logIn(): void
    {
        // Send authentication request to Slack web service.
        // "Logged in to a Slack account '{$this->login}'.";
    }

    public function sendMessage(string $chatId, string $message): void
    {
        // Send message post request to Slack web service.
        // "Posted following message into the '$chatId' chat: '$message'.";
    }
}