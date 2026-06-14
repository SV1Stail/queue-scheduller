# Package Safe

[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg)](http://godoc.mbranch.net/github.com/mbranch/safe-go)

Package safe provides helpers for gracefully handling panics in background
goroutines.

## Example Usage

```golang
import (
  "context"
  "fmt"

  "github.com/mbranch/safe-go"
)

func main() {
  ctx := context.Background()

  items := []int{1, 2, 3, 4, 5}

  g, groupCtx := safe.GroupWithContext(ctx)
  for _, item := range items {
    // Why do we need this wrapper?
    // See: https://medium.com/@julienetienne/golang-for-loop-concurrency-quirk-95e6b184cfe
    func(itemVal int) {
      g.Go(func() error {
        return processItem(groupCtx, itemVal)
      })
    }(item)
  }

  err := g.Wait()
  if err != nil {
    fmt.Println("error", err)
  }
}
```
