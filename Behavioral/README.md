# Behavioral Design Pattern

> Behavioral patterns take care of effective communication and the assignment of responsibilities between objects.

Where creational patterns are about *making* objects and structural patterns about *composing* them, behavioral patterns
are about the traffic between them: who calls whom, in what order, and how much each side needs to know about the other.

| Pattern | One-line intent | Example in this repo |
| --- | --- | --- |
| [Chain of Responsibility](ChainOfResponsibility/README.md) | Pass a request along a chain until someone handles it | HTTP middleware: auth, rate limit, body size |
| [Command](Command/README.md) | Turn a request into an object so it can be queued, logged or undone | A release that rolls itself back on failure |
| [Interpreter](Interpreter/README.md) | Represent a grammar and evaluate sentences in it | A customer-segmentation rule language |
| [Iterator](Iterator/README.md) | Traverse a collection without exposing how it is stored | A cursor over a paginated REST API |
| [Mediator](Mediator/README.md) | Replace a mesh of direct calls with one coordinator | Checkout saga across inventory, payments, shipping |
| [Memento](Memento/README.md) | Snapshot and restore state without exposing internals | Document editor undo/redo |
| [Observer](Observer/README.md) | Notify a changing set of subscribers about events | Order events fanned out to mail, stock, audit |
| [State](State/README.md) | Change behaviour when internal state changes | Order lifecycle with illegal transitions refused |
| [Strategy](Strategy/README.md) | Swap interchangeable algorithms at runtime | Retry backoff policies |
| [Template Method](TemplateMethod/README.md) | Fix an algorithm's skeleton, vary its steps | Nightly CSV/JSON product-feed import |
| [Visitor](Visitor/README.md) | Add operations to a stable object structure | Rendering a document tree to HTML, text, index stats |

## Choosing between the similar ones

These are the pairs that get confused most often:

- **Strategy vs. State** — both delegate to a swappable object. Strategy's alternatives are *independent* and unaware of
  each other; State's know about each other and trigger the transitions between themselves.
- **Strategy vs. Template Method** — Strategy swaps a whole algorithm through composition at runtime; Template Method
  varies individual steps of a fixed skeleton, decided at construction.
- **Chain of Responsibility vs. Decorator** — structurally near-identical. A decorator always passes the call through
  and adds to the result; a chain link may *stop* the request.
- **Command vs. Strategy** — a Command represents a specific request, usually with an undo; a Strategy represents an
  interchangeable way of doing something.
- **Mediator vs. Observer** — Mediator centralises *who talks to whom*; Observer decentralises it by letting anyone
  subscribe. A mediator is often implemented using an observer internally.

## Running the code

Every pattern lives under `<Pattern>/go/` as its own package, with table-driven tests. From the repository root:

```bash
go test ./Behavioral/...
```

If Go is not installed locally:

```bash
docker run --rm -v "$PWD:/app" -w /app golang:1.23-alpine go test ./Behavioral/...
```
