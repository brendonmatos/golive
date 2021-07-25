package main

import (
	"github.com/brendonmatos/golive/live/component"
	"github.com/brendonmatos/golive/live/component/renderer"
)

type Login struct {
	router   *Router
	Email    string
	Password string
}

func (c *Login) DoLogin() {

	//if c.Email != "iii@iii.com" {
	//	return
	//}
	//
	//if c.Password != "123" {
	//	return
	//}
	c.router.Push("/home")
}

func NewLogin() *component.Component {
	c := component.DefineComponent("Login")

	l := &Login{}

	component.OnMounted(c, func(_ *component.Context) {
		l.router = UseRouter(c)
	})

	c.SetState(l)

	err := c.UseRender(renderer.NewTemplateRenderer(`
			<div>
				<span>Login</span>
				<input type="email" gl-input="Email" placeholder="E-mail" />
				<input type="password" gl-input="Password" placeholder="Password" /> 
				<button type="submit" gl-click="DoLogin">Login</button> 

				<br />

				<a href="/register">Register</a> 
			</div>
		`))
	if err != nil {
		return nil
	}

	return c
}
