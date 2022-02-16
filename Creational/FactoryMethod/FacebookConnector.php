<?php

namespace DesignPatterns\Creational\FactoryMethod;

/**
 * This Concrete Product implements the Facebook API.
 */
class FacebookConnector implements SocialNetworkConnector
{
    private string $login;
    private string $password;

    public function __construct(string $login, string $password)
    {
        $this->login = $login;
        $this->password = $password;
    }

    public function logIn(): void
    {
        // todo: Send HTTP API request to log in user $this->login with password $this->password"
    }

    public function logOut(): void
    {
        // todo:  "Send HTTP API request to log out user $this->login";
    }

    public function createPost($content): void
    {
        // todo:  "Send HTTP API requests to create a post in Facebook timeline.";
    }
}