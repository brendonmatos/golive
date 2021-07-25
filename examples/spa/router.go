package main

import (
	"github.com/brendonmatos/golive/live"
	"github.com/brendonmatos/golive/live/component"
	"github.com/brendonmatos/golive/live/component/renderer"
	"github.com/gofiber/fiber/v2"
)

type Router struct {
	Paths     map[string]string
	Path      string
	component *component.Component
}

func (c *Router) ActualRoute() string {
	ctx := component.Inject(c.component, "fiber_ctx").(*fiber.Ctx)
	path := ctx.Path()
	return c.Paths[path]
}

func (c *Router) Push(path string) {
	c.Path = path

	lp := component.Inject(c.component, "page").(*live.Page)

	lp.EmitWithSource(live.PageNavigate, c.component, nil, path)

	// TODO: find some way to send page navigation commands to the browser
	c.component.Update()
}

func ProvideRouter(c *component.Component, router *Router) {
	component.Provide(c, "router", router)
}

func UseRouter(c *component.Component) *Router {
	return component.Inject(c, "router").(*Router)

}

func NewRouter() *component.Component {
	c := component.DefineComponent("Router")

	router := &Router{
		Paths: map[string]string{},
		Path:  "/",
	}

	router.Paths["/"] = "Login"
	router.Paths["/login"] = "Login"
	router.Paths["/home"] = "Home"
	router.Paths["/register"] = "Register"

	ProvideRouter(c, router)
	c.SetState(router)

	c.UseComponent("Login", NewLogin)
	c.UseComponent("Register", NewRegister)
	c.UseComponent("Home", NewHome)

	component.OnBeforeMount(c, func(ctx *component.Context) {

	})

	router.component = c

	err := c.UseRender(renderer.NewTemplateRenderer(`
			<div>
				{{ render .ActualRoute }}
			</div>
		`))
	if err != nil {
		return nil
	}

	return c
}
