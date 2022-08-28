package component

import (
	"errors"
	"fmt"
)

type Context struct {
	Children  []*Context
	Root      *Context
	Closed    bool
	Component *Component
	Hooks
	Provided map[string]interface{}
	Frozen   bool
}

type Hooks map[string][]Hook

type Hook func(ctx *Context)

func NewContext() *Context {
	c := &Context{
		Children: []*Context{},
		Hooks:    Hooks{},
		Provided: map[string]interface{}{},
		Frozen:   false,
		Closed:   false,
	}
	c.Root = c
	return c
}

func (c *Context) Close() error {
	c.Closed = true
	for _, child := range c.Children {
		err := child.Close()
		if err != nil {
			return fmt.Errorf("children ctx close: %w", err)
		}
	}
	return nil
}

func (c *Context) Child() *Context {
	child := NewContext()
	child.Root = c.Root
	c.Children = append(c.Children, child)
	return child
}

func (c *Context) SetHook(targetType string, hook Hook) {
	if c.Frozen {
		return
	}

	if c.Hooks[targetType] == nil {
		c.Hooks[targetType] = []Hook{}
	}

	c.Hooks[targetType] = append(c.Hooks[targetType], hook)
}

func (c *Context) Inherit(i *Context) {
	for t, v := range i.Hooks {
		for _, h := range v {
			c.SetHook(t, h)
		}
	}
}

func (c *Context) InjectGlobalHook(targetType string, hook Hook) {
	r := c.Root
	r.SetHook(targetType, hook)
}

func (c *Context) CallHook(target string) error {
	if c.Closed {
		fmt.Println("---------------------------------------------------------------- CALLING HOOK CLOSED --------------------------------")
		return errors.New("context is closed")
	}

	if hooks, ok := c.Hooks[target]; ok {
		for _, hook := range hooks {
			hook(c)
		}
	}

	r := c.Root

	if r == c {
		return nil
	}

	for _, hook := range r.Hooks[target] {
		hook(c)
	}

	return nil
}
