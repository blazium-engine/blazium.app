# Blazium Services SDK

The Blazium Services SDK is a module specifically designed for seamless integration with the Blazium Engine, enabling effortless interaction with Blazium Services. Unlike traditional HTTP APIs, this SDK is built to feel native to the engine, offering a user-friendly and intuitive approach to incorporating services.

Following our previous article about the [Blazium Services](https://blazium.app/dev-tools/blazium-services) where we described about the services side, this article will focus on the engine side and the client SDK that calls into the services.

### Design Considerations

Since we wanted the SDK to fit right into the engine, we knew we wanted to follow the [best practices when designing modules](https://docs.blazium.app/contributing/development/core_and_modules/custom_modules_in_cpp.html).

### Naming

We named the module blazium_sdk, because it connects to our Blazium Services, and it's an sdk. If in the future we want to add more features to it and it does more than blazium_sdk, we can either split it into multiple modules, or rename it to have a bigger scope.

### Scope

The scope of the module also needs to be well defined, as we wanted to have inside it only things that relate to our Blazium Services. Anything else would need to live in a separate module. This way, if people want to disable the services, they can and still have a working engine.

### Functionality

We wanted every service to be reflected in the engine as a node. Reason is that nodes have lifetime, and we can do things in the constructor of the node, and people can also attach scripts to the nodes if wanted. Singletons would have also fit the design here, however we opted for nodes.

As for using the node, we exposed in the editor both configuration settings (server_url, game_id) and also readonly properties (connected, lobby, peer, peers, host_data, etc.). As such we can both configure the nodes at edit time, but also view at a glance at runtime important information about them.

### API

As for the API, designing a web API can be difficult, especially if you design everything to be with callbacks. We definitely didn't want something like this:

```py
# callback based API
func _ready():
    service.on_connect.connect(_on_connect)
    service.connect()
func _on_connect():
    service.on_do_action.connect(_on_do_action)
    service.do_action()
func _on_do_action():
    ...
```

The main problem with this design is that users would quickly get lost in it, as it goes into callback hell problem. What we did is abstract away this with await, so that people can write instead:

```py
# async based API
func _ready():
    await service.connect().finished
    await service.do_action().finished
    ...
```

This way, everything can be written in one function. The main difference is the returned object would be an object that has a signal on which you can await.

### Happy Flows and Error cases

The service also needs to handle error cases as well as happy flows. In case you call a function and the service is not connected, you don't want to wait forever for an answer. As such, we check inside each function if the sdk is connected, and if its not, returns an object with the finished signal emitted delayed. As for errors, the errors are on the result object.

The SDK is designed to do a network call, and from there return the result. Once you call the SDK, you get a response on which you await for the result.

```py
class Response:
    signal finished(result: Result)

class Result:
    var data: Dictionary
    var error: String

var response: Response = service.call_method()
var result: Result = await response.finished

if result.error != "":
    push_error(result.error)
    return

print(result.data)
```

### Comparison to traditional callback based design

Since everything is designed in mind with awaits, the calls will not block the main thread, and the game will run smoothly until the response is received. Callback based design is similar to this, however it is more error prone, since you might forget to link a callback, or if it fails, the next callback might not be triggered.

Another downside of the callback based design is the difficulty to get the errors out of the calls, since you need a place where to store them, and that complicates the implementation of your game.
