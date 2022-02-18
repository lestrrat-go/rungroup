rungroup
========

Control multiple goroutines from one object.

```go
var g rungroup.Group

for i := 0; i < 10; i++ {
  i := i
  g.Add(rungroup.ActorFunc(func(ctx context.Context) error {
    if i%2 == 1 {
      return fmt.Errorf(`%d`, i)
    }
    return nil
  }))
}

ctx, cancel := context.WithCancel(context.Background())
defer cancel()
err := g.Run(ctx)

time.Sleep(time.Second)
cancel()

if !assert.Len(t, err, 5) {
  return
}
```
