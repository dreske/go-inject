package inject_test

import (
	"github.com/dreske/go-inject"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type SimpleTestInterface interface {
	Test() string
}

type SimpleTestInterfaceImpl struct {
}

func (s *SimpleTestInterfaceImpl) Test() string {
	return "test1"
}

func TestServiceLocator_SimpleBind(t *testing.T) {
	registry := inject.NewRegistry()
	if !assert.NoError(t, registry.Bind("Hello")) {
		return
	}

	result, err := registry.GetByType(reflect.TypeOf(""))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "Hello", result)
}

func TestServiceLocator_BindPointer(t *testing.T) {
	type SimpleTestService struct {
		name string
	}

	registry := inject.NewRegistry()
	if !assert.NoError(t, registry.Bind(&SimpleTestService{name: "test1"})) {
		return
	}

	result, err := registry.GetByType(reflect.TypeOf(&SimpleTestService{}))
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "test1", result.(*SimpleTestService).name)
}

func TestServiceLocator_BindWithName(t *testing.T) {
	registry := inject.NewRegistry()
	if !assert.NoError(t, registry.BindWithName("MyCustomName", "Hello")) {
		return
	}

	result, err := registry.GetByName("MyCustomName", reflect.TypeOf(""))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "Hello", result)
}

func TestServiceLocator_BindWithNameWrongType(t *testing.T) {
	registry := inject.NewRegistry()
	if !assert.NoError(t, registry.BindWithName("MyCustomName", "Hello")) {
		return
	}

	_, err := registry.GetByName("MyCustomName", reflect.TypeOf(1))
	assert.Equal(t, inject.ErrInvalidInjectionType, err)
}

func TestServiceLocator_BindWithType(t *testing.T) {
	registry := inject.NewRegistry()
	err := registry.BindWithType(reflect.TypeOf((*SimpleTestInterface)(nil)).Elem(), &SimpleTestInterfaceImpl{})
	if !assert.NoError(t, err) {
		return
	}

	result, err := registry.GetByType(reflect.TypeOf((*SimpleTestInterface)(nil)).Elem())
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "test1", result.(SimpleTestInterface).Test())
}

func TestServiceLocator_SimpleInject(t *testing.T) {
	registry := inject.NewRegistry()
	if !assert.NoError(t, registry.Bind("Hello")) {
		return
	}

	var test string
	if !assert.NoError(t, registry.Inject(&test)) {
		return
	}
	assert.Equal(t, "Hello", test)
}

func TestServiceLocator_InjectFields(t *testing.T) {
	type Injected struct {
		name string
	}

	type InjectInto struct {
		ServiceByType *Injected `inject:""`
		ServiceByName *Injected `inject:"ServiceByName"`
	}

	registry := inject.NewRegistry()
	if !assert.NoError(t, registry.Bind(&Injected{name: "ServiceByType"})) {
		return
	}
	if !assert.NoError(t, registry.BindWithName("ServiceByName", &Injected{name: "ServiceByName"})) {
		return
	}

	injectInto := InjectInto{}
	if !assert.NoError(t, registry.InjectFields(&injectInto)) {
		return
	}

	if !assert.NotNil(t, injectInto.ServiceByType) {
		return
	}

	assert.Equal(t, "ServiceByType", injectInto.ServiceByType.name)
	assert.Equal(t, "ServiceByName", injectInto.ServiceByName.name)
}

func TestServiceLocator_SimpleInjectInvalidPointer(t *testing.T) {
	registry := inject.NewRegistry()
	if !assert.NoError(t, registry.Bind("Hello")) {
		return
	}

	var test string
	assert.Equal(t, inject.ErrInvalidInjectionPoint, registry.Inject(test))
}

func TestServiceLocator_BindProducer(t *testing.T) {
	registry := inject.NewRegistry()

	err := registry.BindWithType(reflect.TypeOf(""), inject.ProducerFunc(func(source interface{}, target reflect.Type) (interface{}, error) {
		return "Hello World", nil
	}))
	if !assert.NoError(t, err) {
		return
	}

	result, err := registry.GetByType(reflect.TypeOf(""))
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "Hello World", result)
}
