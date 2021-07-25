package main

import (
	"github.com/brendonmatos/golive/live/component"
	"github.com/brendonmatos/golive/live/component/renderer"
)

type Register struct {
	router   *Router
	Email    string
	Password string
}

func (c *Register) DoRegister() {

	if c.Email != "iii@iii.com" {
		return
	}

	if c.Password != "123" {
		return
	}

	c.router.Push("/login")
}

func NewRegister() *component.Component {
	c := component.DefineComponent("Router")

	l := &Register{}

	component.OnMounted(c, func(ctx *component.Context) {
		l.router = UseRouter(c)
	})

	c.SetState(l)

	err := c.UseRender(renderer.NewTemplateRenderer(`
			<div>
				<span>Register</span>
				<input type="email" gl-input="Email" placeholder="E-mail" />
				<input type="password" gl-input="Password" placeholder="Password" /> 
				<button type="submit" gl-click="DoRegister">Register</button> 
				<br />

				<a href="/login">Login</a> 
			</div>
		`))
	if err != nil {
		return nil
	}

	return c
}
