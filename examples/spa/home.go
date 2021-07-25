package main

import (
	"github.com/brendonmatos/golive/live/component"
	"github.com/brendonmatos/golive/live/component/renderer"
)

type Home struct {
	router *Router
}

func (h *Home) Logout() {
	h.router.Push("/login")
}

func NewHome() *component.Component {
	c := component.DefineComponent("Home")
	h := &Home{}

	component.OnMounted(c, func(_ *component.Context) {
		h.router = UseRouter(c)
	})

	c.State.Set(h)

	err := c.UseRender(renderer.NewTemplateRenderer(`
			<div>
				<button gl-click="Logout">Logout</button>
				<span>Home</span>
			</div>
		`))
	if err != nil {
		return nil
	}

	return c
}
