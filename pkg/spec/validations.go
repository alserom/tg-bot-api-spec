package spec

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

type CompositeError struct {
	problems []error
}

func (e *CompositeError) Error() string {
	msg := fmt.Sprintf("%d problems detected:", len(e.problems))
	for _, err := range e.problems {
		msg += "\n- " + err.Error()
	}

	return msg
}

func check(as ApiSpec) error {
	var errs []error

	checks := []func(as ApiSpec, ch chan<- error){
		checkMeta,
		checkDataTypes,
		checkTgTypes,
		checkTgMethods,
	}

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	wg.Add(len(checks))
	for _, job := range checks {
		job := job
		go func() {
			defer wg.Done()
			job(as, ch)
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for err := range ch {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		return &CompositeError{errs}
	}

	return nil
}

func checkMeta(as ApiSpec, ch chan<- error) {
	if as.GetVersion() == "" {
		ch <- errors.New("version not set")
	}

	if as.GetReleaseDate() == "" {
		ch <- errors.New("release date not set")
	}

	if as.GetLink() == "" {
		ch <- errors.New("link not set")
	}
}

func checkDataTypes(as ApiSpec, ch chan<- error) {
	for _, dt := range as.GetDataTypeDefinitions() {
		switch obj := dt.(type) {
		case *ObjectDataType:
			if obj.GetRef() == nil {
				ch <- errors.New("incorrect object data type, reference missed: " + obj.GetDefinition())
			}
		}
	}
}

func checkTgTypes(as ApiSpec, ch chan<- error) {
	for _, t := range as.GetTypes() {
		for _, p := range t.GetProperties() {
			if len(p.GetDataTypes()) == 0 {
				ch <- errors.New(fmt.Sprintf("incorrect property, data type missed (object: %s property: %s", t.GetName(), p.GetName()))
			}
		}

		parent := t.GetParent()
		if parent != nil {
			if _, exists := as.GetType(parent.GetName()); !exists {
				ch <- errors.New(fmt.Sprintf("object %s has a parent %s which is missing in the objects list", t.GetName(), parent.GetName()))
			}
		}

		for _, c := range t.GetChildren() {
			if _, exists := as.GetType(c.GetName()); !exists {
				ch <- errors.New(fmt.Sprintf("object %s has a child %s which is missing in the objects list", t.GetName(), c.GetName()))
			}
		}
	}
}

func checkTgMethods(as ApiSpec, ch chan<- error) {
	for _, m := range as.GetMethods() {
		if len(m.GetReturnTypes()) == 0 {
			ch <- errors.New("return types not set, method: " + m.GetName())
		}

		for _, a := range m.GetArguments() {
			if len(a.GetDataTypes()) == 0 {
				ch <- errors.New(fmt.Sprintf("incorrect argument, data type missed (method: %s argument: %s", m.GetName(), a.GetName()))
			}
		}
	}
}

func validateNonEmptyStringArg(argName, value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(argName + " is required")
	}

	return nil
}

func validateLinkArg(argName, link string) error {
	_, err := url.ParseRequestURI(link)
	if err != nil {
		return errors.New(argName + " is invalid URI for request")
	}

	return nil
}

func skippedAddingNilPoiner() error {
	return errors.New("skipped adding a nil pointer to list")
}
