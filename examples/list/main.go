package main

import (
	"github.com/brendonferreira/golive"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Writter string
type Genre string

type Book struct {
	Writter
	Genre
	Name string
}

type BooksFilter struct {
	Do      bool
	Genre   *Genre
	Writter *Writter
}
type Books struct {
	golive.LiveComponentWrapper
	Filter BooksFilter
	List   []Book
}

func NewBooks() *Books {
	return &Books{
		List: []Book{
			{
				Writter: "Brendon",
				Genre:   "Test",
				Name:    "Go Lang",
			},
			{
				Writter: "Brendon",
				Genre:   "Test",
				Name:    "Go Lang 1",
			},
			{
				Writter: "Brendon",
				Genre:   "Test",
				Name:    "Go Lang 2",
			},
			{
				Writter: "Brendon",
				Genre:   "Test",
				Name:    "Go Lang 3",
			},
		},
		Filter: BooksFilter{
			Do:      true,
			Genre:   nil,
			Writter: nil,
		},
	}
}

func (b *Books) GetFilteredList() []Book {
	filtered := make([]Book, 0)

	for _, book := range b.List {
		match := true
		if b.Filter.Genre != nil && book.Genre != *b.Filter.Genre {
			match = false
		}
		if b.Filter.Writter != nil && book.Writter != *b.Filter.Writter {
			match = false
		}
		if match {
			filtered = append(filtered, book)
		}
	}

	return filtered
}

func (b *Books) ToggleFilter() {
	b.Filter.Do = !b.Filter.Do
}

func NewBooksComponent() *golive.LiveComponent {
	return golive.NewLiveComponent("Books", NewBooks())
}

func (b *Books) TemplateHandler(_ *golive.LiveComponent) string {
	return `
		<div>
			<button go-live-click="ToggleFilter">Toggle</button>

			<div>
				{{ range $index, $Book := .GetFilteredList }}
					<div>
						<span>{{ $Book.Name }}</span>
					</div>
				{{ end }}
			</div>
		</div>	
	`
}

func main() {

	app := fiber.New()
	liveServer := golive.NewServer()

	app.Get("/", liveServer.CreateHTMLHandler(NewBooksComponent, golive.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))

	app.Get("/ws", websocket.New(liveServer.HandleWSRequest))

	_ = app.Listen(":3000")
}
