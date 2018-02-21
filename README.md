# AVURNAVs
Fetch information about AVURNAVs from the different Préfet Maritime websites.

## Example
```go
package main

import (
    "fmt"
    "github.com/antoineaugusti/avurnav"
    "github.com/go-redis/redis"
)

func main() {
    storage := avurnav.NewStorage(redis.NewClient(&redis.Options{
        Addr: ":6379",
    }))
    client := avurnav.NewClient(nil)

    avurnavs, _, err := client.Manche.List()
    if err != nil {
        panic(err)
    }

    fmt.Printf("%+v\n", avurnavs)
    fmt.Printf("%d\n", len(avurnavs))

    for _, avurnav := range avurnavs {
        avurnav, _, _ = client.Manche.Get(avurnav)
        fmt.Println(avurnav, storage.Set(avurnav))
        fmt.Println(storage.Get(avurnav))
    }
}
```
