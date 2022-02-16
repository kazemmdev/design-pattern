# Builder

**Builder** is a creational design pattern that lets you construct complex objects step by step. The pattern allows you
to produce different types and representations of an object using the same construction code.

## Problem

Sometimes we have a class with huge parameters in constructor, or maybe we make the program too complex by creating a
subclass for every possible configuration of an object. here we have builder that suggest the builder class to handle
this situation

## Structure

<img src="assets/scheme.jpg" alt="Builder"/>

## How to Implement

- Make sure that you can clearly define the common construction steps for building all available product
  representations. Otherwise, you won’t be able to proceed with implementing the pattern.
- Declare these steps in the base builder interface.
- Create a concrete builder class for each of the product representations and implement their construction steps.
- Don’t forget about implementing a method for fetching the result of the construction. The reason why this method can’t
  be declared inside the builder interface is that various builders may construct products that don’t have a common
  interface. Therefore, you don’t know what would be the return type for such a method. However, if you’re dealing with
  products from a single hierarchy, the fetching method can be safely added to the base interface.
- Think about creating a director class. It may encapsulate various ways to construct a product using the same builder
  object.
- The client code creates both the builder and the director objects. Before construction starts, the client must pass a
  builder object to the director. Usually, the client does this only once, via parameters of the director’s constructor.
  The director uses the builder object in all further construction. There’s an alternative approach, where the builder
  is passed directly to the construction method of the director.
- The construction result can be obtained directly from the director only if all products follow the same interface.
  Otherwise, the client should fetch the result from the builder.

# Real World Example

One of the best applications of the Builder pattern is an SQL query builder. The builder interface defines the common
steps required to build a generic SQL query. On the other hand, concrete builders, corresponding to different SQL
dialects, implement these steps by returning parts of SQL queries that can be executed in a particular database engine.

<img src="assets/uml.png" alt="Builder Example"/>
