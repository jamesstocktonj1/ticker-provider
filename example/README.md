# Ticker Provider - Counter Example

This example is based off of the [http-keyvalue-counter](https://github.com/wasmCloud/wasmCloud/tree/main/examples/rust/components/http-keyvalue-counter) example found in the official wasmCloud repo. It behaves the same in that it increments a counter in the keyvalue store based off of the path. The way it differes is that it implements the `ticker.Task` function which the `ticker-provider` will call periodically. This function checks for all the keys within the keyvalue store and clears them all.

## Prerequisites
- [wash](https://wasmcloud.com/docs/installation)
- [TinyGo](https://tinygo.org/getting-started/install)
- [Docker](https://docs.docker.com/engine/install)

## Running with wasmCloud

Firstly deploy the keyvalue instance found in the docker compose file using the following command.
```
docker compose up -d
```

You can then build the counter component using wash.
```
wash build -p counter/
```

Deploy the application using wash.
```
wash up -d
wash app deploy wadm.yaml
```

You should then be able to test using curl.
```
$ curl localhost:8080/hello
{
  "count": 1,
  "key": "hello"
}
```

If you keep running this command it will keep on incrementing the value for hello. Then based off of the ticker config, it should reset the counter every 10 seconds causing the counter to start from 0 again.
