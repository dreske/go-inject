package inject

import (
	"errors"
	"github.com/sirupsen/logrus"
	"reflect"
)

var (
	ErrInvalidInjectionPoint = errors.New("invalid injection point")
	ErrEntryNotFound         = errors.New("object not found")
	ErrInvalidInjectionType  = errors.New("invalid injection type")
	ErrFieldNotSettable      = errors.New("field is not settable")
	ErrInvalidProducer       = errors.New("invalid producer")
)

type Producer interface {
	Produce(source interface{}, expectedType reflect.Type) (interface{}, error)
}

type ProducerFunc func(source interface{}, target reflect.Type) (interface{}, error)

func (p ProducerFunc) Produce(source interface{}, target reflect.Type) (interface{}, error) {
	return p(source, target)
}

type Service interface {
	Init(locator *Registry) error
}

type Registry struct {
	log       *logrus.Entry
	populated bool
	entries   map[string]registryEntry
}

type registryEntry struct {
	populated bool
	source    interface{}
}

func NewRegistry() *Registry {
	return &Registry{
		log:       logrus.WithField("module", "Registry"),
		populated: false,
		entries:   make(map[string]registryEntry),
	}
}

func (r *Registry) Bind(service interface{}) error {
	return r.BindWithType(reflect.TypeOf(service), service)
}

func (r *Registry) BindWithType(expectedType reflect.Type, entry interface{}) error {
	actualType := reflect.TypeOf(entry)
	if !r.isAssignableFrom(expectedType, actualType) {
		return ErrInvalidInjectionType
	}
	return r.BindWithName(expectedType.String(), entry)
}

func (r *Registry) BindWithName(name string, entry interface{}) error {
	r.entries[name] = registryEntry{
		populated: false,
		source:    entry,
	}
	return nil
}

func (r *Registry) GetByType(expectedType reflect.Type) (interface{}, error) {
	name := expectedType.String()
	return r.GetByName(name, expectedType)
}

func (r *Registry) getByType(expectedType reflect.Type, source interface{}) (interface{}, error) {
	name := expectedType.String()
	return r.getByName(name, source, expectedType)
}

func (r *Registry) GetByName(name string, expectedType reflect.Type) (interface{}, error) {
	return r.getByName(name, nil, expectedType)
}

func (r *Registry) getByName(name string, source interface{}, expectedType reflect.Type) (interface{}, error) {
	entry, exists := r.entries[name]
	if !exists {
		return nil, ErrEntryNotFound
	}

	actualSource := entry.source
	actualType := reflect.TypeOf(actualSource)
	if actualType != expectedType {
		producer, isProducer := actualSource.(Producer)
		if isProducer {
			producedSource, err := producer.Produce(source, expectedType)
			if err != nil {
				return nil, err
			}

			actualSource = producedSource
			actualType = reflect.TypeOf(actualSource)
		}

		if !r.isAssignableFrom(expectedType, actualType) {
			return nil, ErrInvalidInjectionType
		}
	}

	return actualSource, nil
}

func (r *Registry) isAssignableFrom(expectedType, actualType reflect.Type) bool {
	if expectedType == actualType {
		// actualType is the same as expected
		return true
	}

	if (expectedType.Kind() == reflect.Interface && actualType.Implements(expectedType)) ||
		(expectedType.Kind() == reflect.Ptr && expectedType.Elem().Kind() == reflect.Interface && actualType.Implements(expectedType.Elem())) {
		// an interface is expected and actualType implements it
		return true
	}

	if actualType.Implements(reflect.TypeOf((*Producer)(nil)).Elem()) {
		// actualType is a producer
		return true
	}
	return false
}

func (r *Registry) InjectFrom(caller interface{}, targets ...interface{}) error {
	for _, target := range targets {
		targetPtr := reflect.TypeOf(target)
		if targetPtr.Kind() != reflect.Ptr {
			return ErrInvalidInjectionPoint
		}

		actualValue, err := r.getByName(targetPtr.Elem().String(), caller, targetPtr.Elem())
		if err != nil {
			return err
		}

		serviceValue := reflect.ValueOf(target).Elem()
		if !serviceValue.CanSet() {
			return ErrFieldNotSettable
		}

		serviceValue.Set(reflect.ValueOf(actualValue))
	}
	return nil
}

func (r *Registry) Inject(services ...interface{}) error {
	return r.InjectFrom(nil, services...)
}

func (r *Registry) InjectFields(target interface{}) error {
	targetType := reflect.TypeOf(target).Elem()
	targetValue := reflect.ValueOf(target).Elem()
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		tag, ok := field.Tag.Lookup("inject")
		if !ok {
			continue
		}

		var fieldValue interface{}
		if tag == "" {
			value, err := r.getByType(field.Type, target)
			if err != nil {
				return err
			}
			fieldValue = value
		} else {
			value, err := r.getByName(tag, target, field.Type)
			if err != nil {
				return err
			}
			fieldValue = value
		}

		targetValue.Field(i).Set(reflect.ValueOf(fieldValue))
	}

	return nil
}

func (r *Registry) Populate() error {
	if r.populated {
		r.log.Warn("Service locator is already populated")
		return nil
	}
	for serviceName, entry := range r.entries {
		service, ok := entry.source.(Service)
		if ok {
			r.log.WithFields(logrus.Fields{
				"serviceName": serviceName,
			}).Debug("Populating service")
			if err := service.Init(r); err != nil {
				r.log.WithFields(logrus.Fields{
					"serviceName": serviceName,
				}).WithError(err).Error("Error populating service")
				return err
			}
		}
	}
	return nil
}
