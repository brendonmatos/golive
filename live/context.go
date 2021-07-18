package live

import "fmt"

type Context struct {
	Parent *Context
	Hooks
	Children []*Context
	Root     *Context
	Closed   bool
}

type Hooks map[string][]Hook

type Hook func()

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
	child.Parent = c
	c.Children = append(c.Children, child)
	return c
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
		return nil
	}

	for _, hook := range c.Hooks[target] {
		hook()
	}

	r := c.Root

	if r == c {
		return nil
	}

	return r.CallHook(target)
}
