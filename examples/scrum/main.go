package main

import (
	"fmt"

	"github.com/brendonmatos/golive/impl/live_fiber"
	"github.com/brendonmatos/golive/live"
	"github.com/brendonmatos/golive/live/component"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	app := fiber.New()

	liveFiber := live_fiber.NewFiberServer()

	app.Get("/", liveFiber.CreateLiveComponent(NewBoard, live.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))

	app.Get("/ws", websocket.New(liveFiber.HandleWebSocketConnection))

	_ = app.Listen(":3000")

}

type Column struct {
	Name string
	List []string
}
type Board struct {
	Columns []Column
}

func NewBoard(ctx *component.Context) string {
	state := component.UseState(ctx, &Board{
		Columns: []Column{
			{
				Name: "aaaaa",
				List: make([]string, 0),
			},
			{
				Name: "bbbbb",
				List: make([]string, 0),
			},
		},
	})
	return fmt.Sprintf(`<div>
		<div class="row">
		try! : %s
		</div>
		%s
		<br/>
		%s
	</div>`, state.Columns[0].Name, NewColumn(ctx.Child(), &state.Columns[0]), NewColumn(ctx.Child(), &state.Columns[1]))
}

func NewColumn(ctx *component.Context, props *Column) string {
	state := component.UseState(ctx, props)
	return fmt.Sprintf(`<div>
		<span>Title: %s</span>
		<span>
			<input type="text" value="%s" gl-input="Name" />
		</span>
	</div>`, state.Name, state.Name)
}
