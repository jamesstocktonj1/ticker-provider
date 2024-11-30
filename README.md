# Wasmcloud Ticker Provider

This provider is a simple ticker/cron job caller for [wasmCloud](https://github.com/wasmCloud/wasmCloud). It calls a `ticker.Task` function which is exposed by the component. It is built upon the [go-co-op/gocron/v2](https://github.com/go-co-op/gocron/v2) scheduler.

## Provider Config

This provider can be used in two ways, either a simple interval ticker or as a cron job. Here are some example configurations:

### Interval
```
target_config:
  - name: ticker-config
    properties:
      type: interval        # simple ticker using the `interval` type
      period: 10s           # time period for ticker e.g. 10s, 5m, 1h, etc...
```

### Cron
```
target_config:
  - name: cron-config
    properties:
      type: cron            # cron job using the `cron` type
      cron: "0 * * * *"     # cron configuration e.g. once every hour (every 0th minute)
```
Or
```
target_config:
  - name: cron-config
    properties:
      type: cron            # cron job using the `cron` type
      seconds: "true"       # the `seconds` bool adds an extra column at the start for seconds
      cron: "0 * * * * *"     # cron configuration e.g. once every minute (every 0th second)
```

## Wit Package

In order to use the wit package `jamesstocktonj1:ticker` you must add the namespace to your [wasm-pkg](https://github.com/bytecodealliance/wasm-pkg-tools) config file. To do this run the `wkg config --edit` command and add the following:
```
[namespace_registries]
jamesstocktonj1 = "ghcr.io"
```

You should be able to test whether this works by running the following command:
```
wkg get jamesstocktonj1:ticker
```

You may also need to set the `WASH_PACKAGE_CONFIG_FILE` environment variable with the path to this config file in order for `wash` to be able to pull the right dependencies. On Linux the variable should be:
```
export WASH_PACKAGE_CONFIG_FILE=$HOME/.config/wasm-pkg/config.toml
```

## Example

To use this interface within your component you can add the following to your wit world file:
```
include jamesstocktonj1:ticker/exports@0.1.0;
```

This will export the `jamesstocktonj1:ticker/ticker` interface which the `ticker-provider` links to. Simply implement this interface and the `ticker-provider` will call the `ticker.Task` function on the time interval you have specified. A full component example found in the [example](https://github.com/jamesstocktonj1/ticker-provider/tree/main/example) folder.