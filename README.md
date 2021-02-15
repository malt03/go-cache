# go-cache
go-cache is a thread-safe in-memory cache library that caches a single value in a single instance.  
Unlike key-value stores, it can be implemented without worrying about duplicate keys.

## Installation
```sh
go get github.com/malt03/go-cache
```

## Usage
```go
// Create a cache of 1 hour for ttl and 5 minutes for jitter.
var entitiesCache = cache.New(cache.NewConfig(time.Hour, time.Minute*5))

func GetValues() ([]*Entity, error) {
	entities, err := entitiesCache.Get(func() (interface{}, error) {
		var entities []*Entity
		
		// fetch entities
		
		return entities, nil
	})
	if err != nil {
		return nil, err
	}
	return entities.([]*Entity), nil
}
```

## Other functions
### NoExpiration
```go
var entitiesCache = cache.New(cache.NewConfig(cache.NoExpiration, 0))
```

### Invalidate
```go
entitiesCache.Invalidate()
```
