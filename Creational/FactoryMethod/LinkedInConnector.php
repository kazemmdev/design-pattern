<?php

namespace DesignPatterns\Creational\FactoryMethod;


/**
 * This Concrete Product implements the LinkedIn API.
 */
class LinkedInConnector implements SocialNetworkConnector
{
    private string $email;
    private string $password;

    public function __construct(string $email, string $password)
    {
        $this->email = $email;
        $this->password = $password;
    }

    public function logIn(): void
    {
        // todo:  "Send HTTP API request to log in user $this->email with password $this->password";
    }

    public function logOut(): void
    {
        // todo:  "Send HTTP API request to log out user $this->email";
    }

    public function createPost($content): void
    {
        // todo:  "Send HTTP API requests to create a post in LinkedIn timeline.";
    }
}