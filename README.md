# AVURNAVs
Fetch information about AVURNAVs from the different Préfet Maritime websites.

## Example
```go
package main

import (
    "fmt"
    "github.com/antoineaugusti/avurnav"
)

func main() {
    client := avurnav.NewClient(nil)
    avurnavs, resp, err := client.Atlantique.List()
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", resp)
    fmt.Printf("%+v\n", avurnavs)
}

```
