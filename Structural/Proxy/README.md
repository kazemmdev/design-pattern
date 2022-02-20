# Flyweight

**Proxy** is a structural design pattern that lets you provide a substitute or placeholder for another object. A proxy
controls access to the original object, allowing you to perform something either before or after the request gets
through to the original object.

## Problem

## Structure

<img src="assets/scheme.jpg" alt="Flyweight"/>

## How to Implement

- If there’s no pre-existing service interface, create one to make proxy and service objects interchangeable. Extracting
  the interface from the service class isn’t always possible, because you’d need to change all of the service’s clients
  to use that interface. Plan B is to make the proxy a subclass of the service class, and this way it’ll inherit the
  interface of the service.
- Create the proxy class. It should have a field for storing a reference to the service. Usually, proxies create and
  manage the whole life cycle of their services. On rare occasions, a service is passed to the proxy via a constructor
  by the client.
- Implement the proxy methods according to their purposes. In most cases, after doing some work, the proxy should
  delegate the work to the service object.
- Consider introducing a creation method that decides whether the client gets a proxy or a real service. This can be a
  simple static method in the proxy class or a full-blown factory method.
- Consider implementing lazy initialization for the service object.

# Real World Example

There are countless ways proxies can be used: caching, logging, access control, delayed initialization, etc. This
example demonstrates how the Proxy pattern can improve the performance of a downloader object by caching its results.

<img src="assets/uml.png" alt="Flyweight Example"/>
