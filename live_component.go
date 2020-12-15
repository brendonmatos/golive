package golive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

//
type ComponentLifeTime interface {
	Prepare(component *LiveComponent)
	TemplateHandler(component *LiveComponent) string
	Mounted(component *LiveComponent)
	BeforeMount(component *LiveComponent)
}

type ChildLiveComponent interface{}

//
type LiveComponent struct {
	Name               string
	Component          ComponentLifeTime
	UpdatesChannel     *LifeTimeUpdates
	HTMLTemplateString string
	HTMLTemplate       *template.Template
	Rendered           string
	IsMounted          bool
	Prepared           bool
}

// NewLiveComponent ...
func NewLiveComponent(name string, time ComponentLifeTime) *LiveComponent {
	return &LiveComponent{
		Name:      name,
		Component: time,
	}
}

func (l *LiveComponent) getName() string {
	return l.Name + "_" + NewLiveId().GenerateRandomString()
}

// Prepare 1.
func (l *LiveComponent) Prepare() {
	l.Name = l.getName()

	l.HTMLTemplateString = l.Component.TemplateHandler(l)
	l.HTMLTemplateString = l.addWSConnectScript(l.HTMLTemplateString)
	l.HTMLTemplateString = l.addGoLiveComponentIDAttribute(l.HTMLTemplateString)

	l.HTMLTemplate, _ = template.New(l.Name).Funcs(template.FuncMap{
		"render": func(fn reflect.Value, args ...reflect.Value) (templated template.HTML) {
			child := fn.Interface().(*LiveComponent) // .Call(args)
			child.Mount()
			render := child.Render()
			return template.HTML(render)
		},
	}).Parse(l.HTMLTemplateString)

	l.Component.Prepare(l)

	v := reflect.ValueOf(l.Component).Elem()

	for i := 0; i < v.NumField(); i++ {
		lc, ok := v.Field(i).Interface().(*LiveComponent)

		if !ok {
			continue
		}

		lc.Prepare()
		lc.UpdatesChannel = l.UpdatesChannel

	}

	l.Prepared = true
}

// Mount 2. the component loading html
func (l *LiveComponent) Mount() {
	*l.UpdatesChannel <- ComponentLifeTimeMessage{
		Stage:     WillMount,
		Component: l,
	}

	l.Component.BeforeMount(l)
	l.IsMounted = true

	l.Component.Mounted(l)

	*l.UpdatesChannel <- ComponentLifeTimeMessage{
		Stage:     Mounted,
		Component: l,
	}

}

// GetFieldFromPath ...
func (l *LiveComponent) GetFieldFromPath(path string) *reflect.Value {
	c := (*l).Component
	v := reflect.ValueOf(c).Elem()

	for _, s := range strings.Split(path, ".") {

		// If it`s array this will work
		if i, err := strconv.Atoi(s); err == nil {
			v = v.Index(i)
		} else {
			v = v.FieldByName(s)
		}
	}
	return &v
}

// SetValueInPath ...
func (l *LiveComponent) SetValueInPath(value string, path string) error {
	v := l.GetFieldFromPath(path)
	n := reflect.New(v.Type())

	if v.Kind() == reflect.String {
		value = fmt.Sprintf("\"%s\"", value)
	}

	err := json.Unmarshal([]byte(value), n.Interface())
	if err != nil {
		return err
	}

	v.Set(n.Elem())
	return nil
}

// InvokeMethodInPath ...
func (l *LiveComponent) InvokeMethodInPath(path string, valuePath string) error {
	c := (*l).Component
	v := reflect.ValueOf(c)

	var params []reflect.Value

	if len(valuePath) > 0 {
		params = append(params, *l.GetFieldFromPath(path))
	}

	v.MethodByName(path).Call(params)

	return nil
}

// Render ...
func (l *LiveComponent) Render() string {
	s := bytes.NewBufferString("")
	err := l.HTMLTemplate.Execute(s, l.Component)

	if err != nil {
		fmt.Println(err)
	}

	return s.String()
}

// LiveRender render a new version of the component, and detect
// differences from the last render
// and sets the "new old" version  of render
func (l *LiveComponent) LiveRender() (string, []OutMessage) {
	newRender := l.Render()

	oms := make([]OutMessage, 0)

	if len(l.Rendered) > 0 {

		changeInstructions, err := GetDiffFromRawHTML(l.Rendered, newRender)

		if err != nil {
			panic("There is a error in diff")
		}

		for _, instruction := range changeInstructions {
			oms = append(oms, OutMessage{
				Name:    EventLiveDom,
				Type:    instruction.Type,
				Attr:    instruction.Attr,
				ScopeID: instruction.ScopeID,
				Content: instruction.Content,
				Element: instruction.Element,
			})
		}
	}

	l.Rendered = newRender

	return l.Rendered, oms
}

var re = regexp.MustCompile(`<([a-z0-9]+)`)

func (l *LiveComponent) addWSConnectScript(template string) string {
	return template + `
		<script type="application/javascript">
			goLive.once('WS_CONNECTION_OPEN', function() {
				goLive.connect('` + l.Name + `')
			})
		</script>
	`
}

func (l *LiveComponent) addGoLiveComponentIDAttribute(template string) string {
	found := re.FindString(l.HTMLTemplateString)
	if found != "" {
		replaceWith := found + ` go-live-component-id="` + l.Name + `" `
		template = strings.Replace(l.HTMLTemplateString, found, replaceWith, 1)
	}
	return template
}

// Kill ...
func (l *LiveComponent) Kill() error {

	*l.UpdatesChannel <- ComponentLifeTimeMessage{
		Stage:     WillUnmount,
		Component: l,
	}

	// Set all to nil to garbage collector act
	l.Component = nil
	l.UpdatesChannel = nil
	l.HTMLTemplate = nil

	*l.UpdatesChannel <- ComponentLifeTimeMessage{
		Stage:     Unmounted,
		Component: l,
	}

	return nil
}
