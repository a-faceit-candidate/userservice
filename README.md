# userservice

## Description

`userservice` is a service that provides an implementation of the [`restuser`](https://github.com/a-faceit-candidate/restuser).

This service provides the ability to create, update, delete, retrieve and list users.

This service has MySQL and NSQ as upstream dependencies.

This service exposes basic prometheus metrics for the REST operations it handles.

This service is not intended to be exposed to the internet as it does not handle authentication.

This service exposes a `/status` endpoint for basic healthchecks to be performed.

This service is configurable using environment variables. 
See [`config` struct](./cmd/userservice/main.go) for more details.

## How to run or test this service

Please, install [`golangci-lint`](https://golangci-lint.run/usage/install/#local-installation) and a `docker` daemon to build and test this service.

In order to run all checks run:

```bash
make check
``` 

This will execute both unit and acceptance tests, and will run the lint checks on the code. 

Run `make help` for more detailed targets.

## Acceptance testing

This service uses [_aceptadora_](https://github.com/cabify/aceptadora) to run acceptance tests which relies on the docker image to be previously built.

# Challenge solution

## Introduction

I'm more used to build gRPC APIs as I personally think they RPC APIs are much better for internal backend communication, however, I saw that your company (like the 90% of the internet) provides REST APIs, and I'm also aware of the difficulties of providing gRPC APIs outside of an internal network, so I thought REST would be the best option.
I wanted to provide a library that wouldn't envy anything a gRPC client + protobuf definition provides you, so I decided to build a client that both generates the documentation and serves as an importable library for other services. 

I have to say that I spent more than 50% of the time investigating what's the state of the art on generating REST clients from golang documentation, or documentation from golang code, and I was really disappointed on the lack of progress on this topic.
The two most popular libraries that generate Golang code from an OpenAPI specification are either incomplete (and would require filling them manually) or tied to a specific HTTP libraries, so what I did was to build a client that uses a standard `http.Client` as the transport (we could improve that by accepting any kind of `Doer` actually) and I used the client method annotations to generate the documentation.

## Libraries and dependencies

I used `gin` as router, as it's the one I'm most familiar with (with both good and bad things).

I used MySQL as persistence solution, and NSQ as our event transport.

I used my fork of `envconfig` library to parse the envconfig, although in this service we don't use the slices of structs so we could use the original one.

I used `go-sqlbuilder` for SQL query building. I prefer to avoid heavy ORMs as having more control on the persistence layer is important for a service with high throughput.

I used `logrus` for logging, being abstracted by the internal `log` package that also manages the context logging feature. Changing to another logger shouldn't require changes in any package except for the `log` and `main.go`.

Finally, I use prometheus for metrics. I didn't include a lot of them, more on this later, if were adding them, I'd probably use [`gotoprom`](https://github.com/cabify/gotoprom) for defining them.

## Service structure

This service has three main parts: 
- the `api` (controller) that maps to the business model and actions on the `service` package.
- the `service` package containing our business logic
- the `persistence` package that depends on the model and communicates with the persistence layer.

Additionally we have the `event` package which is an implementation of the Observer pattern for event publishing. 
In our case the event system is NSQ.

There's a small overhead on mapping the models and errors between the different layers, however that overhead is paid by the simplicity of future changes, like the transport change (even a major versioning) and keeping the logic in its right place also follows the law of the least surprise.

## Events

A lot of decisions and assumptions were taken on events, but TL;DR, they don't offer a delivery guarantee right now. 
The transport used is NSQ, which ensures delivery unless the node is lost, as it doesn't provide high availability.
The code was written with the same decision, and we send the event _after_ performing the CRUD operations. 
A more complex solution ensuring delivery would require a lot more code and is IMO, out of scope of a code challenge.
Finally, we decided to publish in the events just the ID of the user that was affected by an operation.
Usually an event system would require its own model, or maybe it could reuse the common REST model for that, but I didn't see the need of spending time on implementing that.

## Testability

We all care about tests, I do too. 
But I didn't see the value in writing hundreds of lines of code that nobody would ever read, so I didn't test everything. 
For example, I did tests for the small `event.NSQPublisher`, but I only wrote one testcase for `service.ServiceImpl` and then I left you some funny ascii art to look at. 
I hope you'll like it.
I did no tests at all for the `api.UsersResource`, but I do think they're needed. 
I'd write those starting a `httptest.Server` with a real `gin` and would use the `restuser` client, asserting expectations on the `servicemock.Service` mock (that's why the mocks are generated but they're not used).

An interesting point are the acceptance tests you can find in the `acceptance/` submodule.
They use the `aceptadora` library I wrote for my current job (we've open-sourced it recently) and they cover most of the functionality of the service. 
This kind of tests remove most of the manual-testing and are the most valuable on the long run.

My personal opinion that the only _good_ way to test is to use a combination of both. 
Unit tests are not enough, but they're an efficient way of checking that weird case a developer knows that can happen on some path, a good example for this is the duplicate row test for the MySQL repository implementation.
Acceptance tests are great, but they can't cover some scenarios (mostly error handling) in a reasonable amount of code.

A special mention is to the MySQL test, which often people assert that the queries the ORM/sql-builder are the right ones: if I wrote those, there's a 1000 to 1 chance that I'd be wrong writing the query in the assertion than the sql builder doing it wrong: the best way of testing that is to test it against a real MySQL. 
I wish there was such a thing as [`miniredis`](https://github.com/alicebob/miniredis) for MySQL, but there isn't and here's where the acceptance tests come to the rescue. 
However, it _is_ important to test things like error handling in the repository implementation.

## Scalability

The provided implemenation will scale until it doesn't. 
The first thing we should probably do for a real production usecase is to add database sharding here, however, that would require additional storage for the indexes. 
Depending on the consistency requirements of those indexes, we could use the same or other storage or even service for them.
We should probably also add caching, using a `memcached` or an in-memory solution (we could use NSQ messages for invalidation), but this would also depend on our consistency requirements.
The service itself can scale horizontally with no limits.

## Misc decisions

You'll find a lot of pointers that can raise you concerns of mutability of garbage collection, however, I'd say that justifying my decision would take another readme like this. Lets just keep it consistent everywhere for now.

Lack of metrics: I'd love to add metrics to MySQL repository, for instance, however, I'd rather add them to a common client wrapper to keep those metrics consistent across all our services. I'd also add some circuit breakers to that client too. Same happens with the NSQ client. 

`service.ServiceImpl` could also have some metrics, however in this small service they wouldn't provide more information than the immediately previous transport layer.

## Password handling

This topic deserves a special mention. Password handling is weird in this service (for instance, we wouldn't require the password on every update right? 
OTOH, other services that just want to render a username, retrieve the password hash/salt as a side effect). 
Definitely this can be improved, but in general terms this is a smell.

In my opinion, this authentication details should be stored in a completely different service (a user can have a password or it may not have one, in 2020 there are tons of alternatives for authentication).
