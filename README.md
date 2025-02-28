# Cache

Cache code is located in the `cache` package. See an example execution with `go run cmd/example/main.go`.

```
//Run tests with coverage
go test -coverprofile=coverage.out ./...

//View coverage report
go tool cover -html=coverage.out
```

## Implementation
This in-memory cache stores key/value pairs and supports expiration. When a key expires, it is no longer returned. However it will not be evicted until the cache is at capacity and needs space or it is manually pruned. This separation of expiration and eviction allows the cache to operate without any supervisory process performing cleanup. This makes it easier to reason about and write tests for edge cases.

For tracking expiration it holds a separate reference to all cache entries in a priority queue. This minimizes the time finding an expired entry when at capacity and provides evicient updates to keys that change the expiration.

For safe concurrent usage, it relies on mutexes, locking around the cache and expiration data structures.

`cmd/example` starts two goroutines which set and fetch items from the queue. The root `img` package holds shared types/interfaces that aren't specific to this cache implementation.

## Future
To take this further there are changes needed to productionalize it but also potential refactors and extentions depending on real world usage.

* Decouple keys/values from string/string to allow more general caching
* Structured logging
* Emit stats as metrics
* Consider tracking least recently used (LRU) key for eviction when at capacity and no keys have expired
* Extend API to include a context to allow logging to be scoped by the calling operation 
* Better separation between cache implementation and heap
* Consider a cache-filling mechanism if use-cases could benefit
