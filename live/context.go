package live

import "fmt"

type Context struct {
	Parent *Context
	Hooks
	Children []*Context
	Global   *Context
}

type Hooks map[string][]Hook

type Hook func()

func NewContext() *Context {
	return &Context{Children: []*Context{}, Hooks: Hooks{}}
}

func (c *Context) Child() *Context {
	ctx := NewContext()
	ctx.Parent = c
	c.Children = append(c.Children, ctx)
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

	r := c.GetRoot()

	if r.Global == nil {
		r.Global = NewContext()
	}

	g := r.Global

	g.InjectHook(targetType, hook)
}

func (c *Context) GetRoot() *Context {
	w := c
	for {
		if c.Parent == nil {
			break
		}

		w = c.Parent
	}

	return w
}

func (c *Context) CallHook(target string) error {
	fmt.Println("calling", target, len(c.Hooks))
	if c.Hooks[target] == nil {
		fmt.Println("hooks undefined", target)
		return nil
	}

	for _, hook := range c.Hooks[target] {
		hook()
	}

	g := c.GetRoot().Global

	if g != nil {
		g.CallHook(target)
	}

	return nil
}
