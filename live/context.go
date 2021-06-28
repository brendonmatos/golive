package live

type Context struct {
	Pairs map[string]interface{}
}

func NewContext() Context {
	return Context{
		Pairs: map[string]interface{}{},
	}
}
