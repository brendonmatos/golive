package main

import (
	"github.com/brendonmatos/golive/live"
	"github.com/brendonmatos/golive/live/component"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Writer string
type Genre string

type Book struct {
	Writer
	Genre
	Name string
}

type BooksFilter struct {
	Do bool
	Writer
	Genre
}

type Books struct {
	component.Wrapper
	Filter BooksFilter
	List   []Book
}

func NewBooks() *Books {
	return &Books{
		List: []Book{
			{
				Writer: "J. K. Rowling",
				Genre:  "fantasy",
				Name:   "Harry Potter and the Philosopher's Stone",
			},
			{
				Writer: "Caleb Doxsey",
				Genre:  "programming",
				Name:   "Introducing Go: Build Reliable, Scalable Programs",
			},
			{
				Writer: "Marijn Haverbeke",
				Genre:  "programming",
				Name:   "Eloquent JavaScript: A Modern Introduction to Programming",
			},
		},
		Filter: BooksFilter{
			Do:     true,
			Genre:  "",
			Writer: "",
		},
	}
}

func NewBooksComponent() *component.Component {
	return component.NewLiveComponent("Books", NewBooks())
}
func (b *Books) GetFilteredList() []Book {
	filtered := make([]Book, 0)

	for _, book := range b.List {
		match := true
		if b.Filter.Genre != "" && book.Genre != b.Filter.Genre {
			match = false
		}
		if b.Filter.Writer != "" && book.Writer != b.Filter.Writer {
			match = false
		}
		if match {
			filtered = append(filtered, book)
		}
	}

	return filtered
}

func (b *Books) SetFilterWriter(data map[string]string) {
	b.Filter.Do = !b.Filter.Do

	if name, ok := data["writer"]; ok {
		b.Filter.Writer = Writer(name)
	}
}

func (b *Books) GetWriters() []Writer {
	writers := make([]Writer, 0)

book:
	for _, book := range b.List {
		for _, writer := range writers {
			if writer == book.Writer {
				continue book
			}
		}
		writers = append(writers, book.Writer)
	}
	return writers
}

func (b *Books) TemplateHandler(_ *component.Component) string {
	return `
		<div>
			<select gl-input="Filter.Writer">
				<option value="">No Filter</option>
				{{ range $index, $writer := .GetWriters }}
					<option value="{{$writer}}">{{$writer}}</option>
				{{ end }}
			</select>
					
			<div>
				{{ range $index, $Book := .GetFilteredList }}
					<div style="border: 1px solid black; padding: 10px;" key="{{$index}}">
						<span><b>name:</b> {{ $Book.name }}</span><br>
						<span><b>Writer:</b> {{ $Book.Writer }}</span>
					</div>
				{{ end }}
			</div>
		</div>	
	`
}

func main() {

	app := fiber.New()
	liveServer := live.NewServer()

	app.Get("/", liveServer.CreateStaticPageRender(NewBooksComponent, live.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))

	app.Get("/ws", websocket.New(liveServer.HandleWebSocketConnection))

	_ = app.Listen(":3000")
}
