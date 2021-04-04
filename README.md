# go-inject

Simple service registry used in some of my private projects.

## Usage
### Binding simple values (like config strings)
```go
registry := inject.NewRegistry()
registry.BindWithName("MyConfigValue", "Hello")

result, err := registry.GetByName("MyConfigValue", reflect.TypeOf(""))
```

### Binding structs (e.g. services)
```go
registry := inject.NewRegistry()
registry.Bind(&SimpleTestService{})

result, err := registry.GetByType(reflect.TypeOf(&SimpleTestService{}))
```

### Injecting (manual)
After binding all required services to the registry, call
```go
registry.Populate()
```

`Init()` function will be called on all bindings implementing `inject.Service`interface.
```go
type MyService struct {
	log *logrus.Entry
}

func (m *MyService) Init(registry *inject.Registry) error {
	// this lookup an entry for type *logrus.Entry in the registry an set it
    return registry.InjectFrom(m, &m.log)
}
```

### Injecting (automatically)
After binding all required services to the registry, call
```go
registry.Populate()
```

The registry will look for annotated fields in all registered entries and inject the appropriate object.
```go
type InjectInto struct {
    ServiceByType *Injected `inject:""`
    ServiceByName *Injected `inject:"ServiceByName"`
}   
```

### Producers

Producer structs or methods that implement the `inject.Producer` interface.
They are creating the value to inject in the moment of injection.

This gives the ability of creating a specialized instance for each injection.
```go
registry.BindWithType(reflect.TypeOf(&logrus.Entry), inject.ProducerFunc(func(source interface{}, target reflect.Type) (interface{}, error) {
    sourceType := reflect.TypeOf(source)
    if sourceType.Kind() == reflect.Ptr {
        sourceType = sourceType.Elem()
    }
    return logrus.WithField("module", sourceType.Name()), nil
}))
```

```go
type ProducerDemo struct {
    log *logrus.Entry `inject:""`
}   
```