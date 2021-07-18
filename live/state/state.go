package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brendonmatos/golive/live/util"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrStateCannotResolvePath = errors.New("cannot resolve path in state")
)

type State struct {
	Value interface{}
}

func NewState() *State {
	return &State{}
}

func (s *State) Set(c interface{}) {
	s.Value = c
}

func (s *State) GetFieldFromPath(path string) (*reflect.Value, error) {
	c := (*s).Value
	v := reflect.ValueOf(c).Elem()

	for _, s := range strings.Split(path, ".") {

		if reflect.ValueOf(v).IsZero() {
			return nil, ErrStateCannotResolvePath
		}

		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		// If it`s array this will work
		if i, err := strconv.Atoi(s); err == nil {
			v = v.Index(i)
		} else {
			v = v.FieldByName(s)
		}
	}
	return &v, nil
}

func (s *State) InvokeMethodInPath(path string, args []reflect.Value) ([]reflect.Value, error) {
	m := reflect.ValueOf(s.Value).MethodByName(path)

	if !m.IsValid() {
		return nil, fmt.Errorf("not a valid function: %v", path)
	}

	// TODO: check for errors when calling
	switch m.Type().NumIn() {
	case 0:
		return m.Call(nil), nil
	case 1:
		return m.Call(
			[]reflect.Value{args[0]},
		), nil
	case 2:
		return m.Call(args), nil
	}

	return nil, nil
}

func (s *State) SetValueInPath(value string, path string) error {

	sf, _ := s.GetFieldFromPath(path)

	if sf.Kind() == reflect.Ptr {
		f := sf.Elem()
		sf = &f
	}

	n := reflect.New(sf.Type())

	if sf.Kind() == reflect.String {
		value = `"` + util.JsonEscape(value) + `"`
	}

	err := json.Unmarshal([]byte(value), n.Interface())
	if err != nil {
		return err
	}

	sf.Set(n.Elem())

	return nil
}

func (s *State) Kill() {
	s.Set(nil)
}
