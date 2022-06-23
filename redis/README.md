# Package redis

## Description

Package redis is a package containing shared utility functions and structs used by Stori Golang resources. The Redis package uses [GO-REDIS](https://github.com/go-redis/redis)

The `NewConn` function return the Redis connection struct that is used to call Redis functions
The purpose of this package is to standardize the way we use Redis functions .

### Example

Import Package

```go
package main

import (
    "fmt"
    "github.com/credifranco/stori-utils-go/redis"
)

func main() {

    // Pass host, port and password as parameter
    rConn,err = redis.NewConn("localhost","6379","")
    if err =nil {
        fmt.fatal(err.Error())
    }
    // Set the new Key with expiration 
    if err := rConn.Set("new_test_key","val_123",30*time.Minutes); err != nil {
        fmt.fatal(err.Error())
    }

    // Get the key Value
    data, err = rConn.Get("new_test_key")
    if err != nil {
        fmt.fatal(err.Error())
    }

    fmt.Println(data)
}

```
