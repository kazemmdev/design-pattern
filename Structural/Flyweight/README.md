# Flyweight

**Flyweight** is a structural design pattern that lets you fit more objects into the available amount of RAM by sharing
common parts of state between multiple objects instead of keeping all the data in each object.

## Problem

Imagine that you must make your code work with a broad set of objects that belong to a sophisticated library or
framework. Ordinarily, you’d need to initialize all of those objects, keep track of dependencies, execute methods in the
correct order, and so on.

As a result, the business logic of your classes would become tightly coupled to the implementation details of 3rd-party
classes, making it hard to comprehend and maintain.

> Note that real applications for the Flyweight pattern in PHP are pretty scarce. This stems from the single-thread
> nature of PHP, where you’re not supposed to be storing ALL of your application’s objects in memory at the same time,
> in the same thread.

## Structure

<img src="assets/scheme.jpg" alt="Flyweight"/>

## How to Implement

- Divide fields of a class that will become a flyweight into two parts:
    - the intrinsic state: the fields that contain unchanging data duplicated across many objects the extrinsic state:
      the fields that contain contextual data unique to each object Leave the fields that represent the intrinsic state
      in the class, but make sure they’re immutable. They should take their initial values only inside the constructor.
    - Go over methods that use fields of the extrinsic state. For each field used in the method, introduce a new
      parameter and use it instead of the field.

- Optionally, create a factory class to manage the pool of flyweights. It should check for an existing flyweight before
  creating a new one. Once the factory is in place, clients must only request flyweights through it. They should
  describe the desired flyweight by passing its intrinsic state to the factory.
- The client must store or calculate values of the extrinsic state (context) to be able to call methods of flyweight
  objects. For the sake of convenience, the extrinsic state along with the flyweight-referencing field may be moved to a
  separate context class.

# Real World Example

In this example, the **Flyweight** pattern is used to minimize the RAM usage of objects in an animal database of a
cat-only veterinary clinic. Each record in the database is represented by a Cat object. Its data consists of two parts:

- Unique (extrinsic) data such as a pet’s name, age, and owner info.
- Shared (intrinsic) data such as a breed name, color, texture, etc.

The first part is stored directly inside the Cat class, which acts as a context. The second part, however, is stored
separately and can be shared by multiple cats. This shareable data resides inside the CatVariation class. All cats that
have similar features are linked to the same CatVariation class, instead of storing the duplicate data in each of their
objects.

<img src="assets/uml.png" alt="Flyweight Example"/>
