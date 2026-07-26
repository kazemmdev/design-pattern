# Design Pattern

Design patterns are typical solutions to commonly occurring problems in software design. They are like pre-made
blueprints that you can customize to solve a recurring design problem in your code.

Each pattern has its own folder containing a `README.md` explaining the problem it solves, and one subfolder per
language holding a runnable implementation with tests.

## Creational Design Pattern

> patterns provide object creation mechanisms that increase flexibility and reuse of existing code.
>  - [Factory Method](Creational/FactoryMethod/README.md)
>  - [Abstract Factory](Creational/AbstractFactory/README.md)
>  - [Builder](Creational/Builder/README.md)
>  - [Prototype](Creational/Prototype/README.md)
>  - [Singleton](Creational/Singleton/README.md)

## Structural Design Pattern

> patterns explain how to assemble objects and classes into larger structures, while keeping the structures flexible and
> efficient.
>  - [Adapter](Structural/Adapter/README.md)
>  - [Bridge](Structural/Bridge/README.md)
>  - [Composite](Structural/Composite/README.md)
>  - [Decorator](Structural/Decorator/README.md)
>  - [Facade](Structural/Facade/README.md)
>  - [Flyweight](Structural/Flyweight/README.md)
>  - [Proxy](Structural/Proxy/README.md)

## Behavioral Design Pattern

> patterns take care of effective communication and the assignment of responsibilities between objects.
>  - [Chain of Responsibility](Behavioral/ChainOfResponsibility/README.md)
>  - [Command](Behavioral/Command/README.md)
>  - [Interpreter](Behavioral/Interpreter/README.md)
>  - [Iterator](Behavioral/Iterator/README.md)
>  - [Mediator](Behavioral/Mediator/README.md)
>  - [Memento](Behavioral/Memento/README.md)
>  - [Observer](Behavioral/Observer/README.md)
>  - [State](Behavioral/State/README.md)
>  - [Strategy](Behavioral/Strategy/README.md)
>  - [Template Method](Behavioral/TemplateMethod/README.md)
>  - [Visitor](Behavioral/Visitor/README.md)

See [Behavioral/README.md](Behavioral/README.md) for a comparison of the patterns that are easiest to confuse with each
other — Strategy vs. State, Mediator vs. Observer, and so on.

## Layout

```text
<Category>/<Pattern>/
├── README.md        problem, structure, how to implement, real-world example
├── assets/          diagrams (Creational and Structural)
├── php/             PHP implementation
│   └── Tests/
└── go/              Go implementation, with _test.go alongside
```

## Running the code

### Go

The Behavioral patterns are implemented in Go, as one package per pattern under a single module.

```bash
go test ./...
```

Go is not required locally — the toolchain runs fine in a container:

```bash
docker run --rm -v "$PWD:/app" -w /app golang:1.23-alpine go test ./...
```

### PHP

The Creational and Structural patterns are implemented in PHP and tested with PHPUnit.

```bash
composer install
composer test
```

## Status

| Category | Patterns | Language |
| --- | --- | --- |
| Creational | 5 / 5 | PHP |
| Structural | 7 / 7 | PHP |
| Behavioral | 11 / 11 | Go |

Go implementations of the Creational and Structural patterns, and PHP implementations of the Behavioral ones, are not
written yet — the `php/` and `go/` layout is in place for both.
