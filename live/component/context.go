package component

import (
	"fmt"
)

type Context struct {
	Children  []*Context
	Root      *Context
	Closed    bool
	Component *Component
	Hooks
}

type Hooks map[string][]Hook

type Hook func(ctx *Context)

func NewContext() *Context {
	c := &Context{Children: []*Context{}, Hooks: Hooks{}}
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

func (c *Context) InjectHook(targetType string, hook Hook) {
	if c.Hooks[targetType] == nil {
		c.Hooks[targetType] = []Hook{}
	}

	c.Hooks[targetType] = append(c.Hooks[targetType], hook)
}

func (c *Context) Inherit(i *Context) {
	for t, v := range i.Hooks {
		for _, h := range v {
			c.InjectHook(t, h)
		}
	}
}

func (c *Context) InjectGlobalHook(targetType string, hook Hook) {
	r := c.Root
	r.InjectHook(targetType, hook)
}

func (c *Context) CallHook(target string) error {

	if c.Hooks[target] == nil {
		for _, hook := range c.Hooks[target] {
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
