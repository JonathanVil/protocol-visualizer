# System Basics

## Steps
The simulator is built on a "step" system.
A step is executed every x seconds, where x is the stepsize, which can be modified through the UI.

## Messages
Every step, moves each message currently in transit one step closer to its recipient.
The amount of steps in transit depends on the transit time, which may be modified through the UI.
The actual transit time of any message will be some random value between the `MIN` and `MAX`.
Once a message reaches its target, the `receive` method of the target will be called.

## Manually sending messages
You can manually send a message from one actor to another by specifying the sender id, target id, type, and data of the message, and pressing send.


# Implementation Basics

The first step in simulating a protocol is the implementation of the actors.
A correct implementation must have several qualities.
It must provide a class specification called `Actor`, and a constructor for this class. It must also implement a `receive(msg)` method.
Lastly, it must return the class at the end of the program.
Provided below is a bare minimum implementation:
```js
class Actor {
    constructor(id) {
        this.id = id;
    }
    receive(msg) {
        if (msg.type === x) {
            //do something
        }
    }
}
return Actor
```

## Fields
An actor must have an `id` field. This allows other actors to send messages to it. As long as you add the field, the simulation will automatically assign a unique id.
Other fields may be added in the constructor.
Setting the `nodeColor` field like: `this.nodeColor = '#1d4ed8'` allows you to customize the color of a node that is used in the visualization.

## Receive
The receive method should match on every type of message you expect an actor to receive.

For inspiration on how to implement a protocol yourself, you can load in some pre-implemented protocols by using the dropdown menu at the top of the screen.


# Built-in functions

## `send(from, to, type, ?data)`
This function allows sending messages between nodes.

It takes four arguments:
- `from` should always be the sender's ID, which can be accessed with `this.id`. This enables the recipient to reply if necessary.
- `to` is the recipient's ID.
- `type` is a string that specifies the kind of message being sent. In practice, it is used for matching messages in the `receive` method.
- `data` is an optional field that allows you to send an object containing additional information, such as timestamps.

## `getActors()`
returns the number of actors currently in the systems

## `createQueue()`
returns an empty linked list queue

### Queue functions
- `queue.push(element)` allows you to push one element to the back of the queue.
- `queue.pop()` allows you tot pop one element from the front of the queue.
- `queue.length` gives the amount of elements in the queue.

## `timeout(actor, steps, reaction)`
This function allows calling a method after a set amount of steps.

It takes three arguments:
- `actor` should always be `this`. This is important to ensure the method can access state.
- `steps` is the number of steps the timer waits before executing the reaction.
- `reaction` is a reference to the method to be called when the timer finishes. This method must take no arguments.