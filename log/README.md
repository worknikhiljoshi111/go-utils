# Package log

## Description
This package contains logging functions to create a new JSON logger. The log function use the
[ZAP](https://github.com/uber-go/zap) package under the hood.


The `NewLogger` function returns a `*zap.SugaredLogger` pointer. Check out the
[ZAP](https://github.com/uber-go/zap) package for more info on how to use this.

When creating a new logger, you can use an environment variable to control the log level. The
variable `LOG_LEVEL` must be set to one of the following.
- `DEBUG`
- `INFO`
- `WARN`
- `ERROR`

## Example
```go
package main

import "github.com/credifranco/stori-utils-go/log"

func main() {
    // LOG_LEVEL = DEBUG
    logger, err := log.NewLogger()
    logger.Debugw("example", "key", "value")
}
