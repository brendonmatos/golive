package golive

import (
	"bytes"
	"encoding/json"
	"errors"
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
	BeforeUnmount(component *LiveComponent)
}

type ChildLiveComponent interface{}

//
type LiveComponent struct {
	Name               string
	component          ComponentLifeTime
	updatesChannel     *ComponentLifeCycle
	htmlTemplateString string
	htmlTemplate       *template.Template
	rendered           string
	IsMounted          bool
	Prepared           bool
	Exited             bool
	log                Log
}

// NewLiveComponent ...
func NewLiveComponent(name string, time ComponentLifeTime) *LiveComponent {
	return &LiveComponent{
		Name:      name,
		component: time,
	}
}

func (l *LiveComponent) getName() string {
	return l.Name + "_" + NewLiveId().GenerateSmall()
}

func (l *LiveComponent) RenderChild(fn reflect.Value, _ ...reflect.Value) template.HTML {
	child, ok := fn.Interface().(*LiveComponent)
	if !ok {
		l.log(LogError, "child not a *golive.LiveComponent", nil)

		return ""
	}

	child.Mount()

	render, err := child.Render()
	if err != nil {
		l.log(LogError, "render child: render", logEx{"error": err})
	}

	return template.HTML(render)
}

// Prepare 1.
func (l *LiveComponent) Prepare() {
	l.log(LogTrace, "Prepare", logEx{"name": l.Name})

	l.Name = l.getName()

	l.htmlTemplateString = l.component.TemplateHandler(l)
	l.htmlTemplateString = l.addWSConnectScript(l.htmlTemplateString)
	l.htmlTemplateString = l.addGoLiveComponentIDAttribute(l.htmlTemplateString)

	l.htmlTemplate, _ = template.New(l.Name).Funcs(template.FuncMap{
		"render": l.RenderChild,
	}).Parse(l.htmlTemplateString)

	l.component.Prepare(l)

	l.PrepareChildren()

	l.Prepared = true
}

func (l *LiveComponent) PrepareChildren() {
	v := reflect.ValueOf(l.component).Elem()
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanInterface() {
			continue
		}

		lc, ok := v.Field(i).Interface().(*LiveComponent)

		if !ok {
			continue
		}

		lc.updatesChannel = l.updatesChannel
		lc.log = l.log
		lc.Prepare()
	}
}

func (l *LiveComponent) KillChildren() {
	v := reflect.ValueOf(l.component).Elem()
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanInterface() {
			continue
		}

		lc, ok := v.Field(i).Interface().(*LiveComponent)

		if !ok {
			continue
		}

		if err := lc.Kill(); err != nil {
			l.log(LogError, "kill child", logEx{"name": lc.Name})
		}
	}
}

// Mount 2. the component loading html
func (l *LiveComponent) Mount() {
	l.log(LogTrace, "WillMount", logEx{"name": l.Name})

	*l.updatesChannel <- ComponentLifeTimeMessage{
		Stage:     WillMount,
		Component: l,
	}

	l.component.BeforeMount(l)
	l.IsMounted = true

	l.component.Mounted(l)

	l.log(LogTrace, "Mounted", logEx{"name": l.Name})

	*l.updatesChannel <- ComponentLifeTimeMessage{
		Stage:     Mounted,
		Component: l,
	}
}

// GetFieldFromPath ...
func (l *LiveComponent) GetFieldFromPath(path string) *reflect.Value {
	c := (*l).component
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
func (l *LiveComponent) InvokeMethodInPath(path string, data map[string]string, domEvent *DOMEvent) error {
	m := reflect.ValueOf(l.component).MethodByName(path)
	if !m.IsValid() {
		return fmt.Errorf("not a valid function: %v", path)
	}

	// TODO: check for errors when calling
	switch m.Type().NumIn() {
	case 0:
		m.Call(nil)
	case 1:
		m.Call(
			[]reflect.Value{reflect.ValueOf(data)},
		)
	case 2:
		m.Call(
			[]reflect.Value{
				reflect.ValueOf(data),
				reflect.ValueOf(domEvent),
			},
		)
	}

	return nil
}

// Render ...
func (l *LiveComponent) Render() (string, error) {
	l.log(LogTrace, "Render", logEx{"name": l.Name})

	if l.component == nil {
		return "", errors.New("component nil")
	}

	s := bytes.NewBufferString("")

	err := l.htmlTemplate.Execute(s, l.component)

	if err != nil {
		return "", err
	}

	return s.String(), nil
}

// LiveRender render a new version of the component, and detect
// differences from the last render
// and sets the "new old" version  of render
func (l *LiveComponent) LiveRender() ([]OutMessage, error) {
	newRender, err := l.Render()

	if err != nil {
		return nil, fmt.Errorf("render component: %w", err)
	}

	oms := make([]OutMessage, 0)
	if len(l.rendered) > 0 {

		changeInstructions, err := GetDiffFromRawHTML(l.rendered, newRender)

		if err != nil {
			l.log(LogPanic, "there is a error in diff", logEx{"error": err})
		}

		for _, instruction := range changeInstructions {
			oms = append(oms, OutMessage{
				Name:        EventLiveDom,
				Type:        strconv.Itoa(int(instruction.Type)),
				Attr:        instruction.Attr,
				ComponentId: instruction.componentId,
				Content:     instruction.Content,
				Element:     instruction.Element,
			})
		}
	}

	l.rendered = newRender

	return oms, nil
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

// TODO: improve this urgently
func (l *LiveComponent) addGoLiveComponentIDAttribute(template string) string {
	found := re.FindString(l.htmlTemplateString)
	if found != "" {
		replaceWith := found + ` go-live-component-id="` + l.Name + `" `
		template = strings.Replace(l.htmlTemplateString, found, replaceWith, 1)
	}
	return template
}

// Kill ...
func (l *LiveComponent) Kill() error {
	l.KillChildren()

	l.log(LogTrace, "WillUnmount", logEx{"name": l.Name})

	l.component.BeforeUnmount(l)

	*l.updatesChannel <- ComponentLifeTimeMessage{
		Stage:     WillUnmount,
		Component: l,
	}

	l.Exited = true
	// Set all to nil to garbage collector act
	l.component = nil
	l.htmlTemplate = nil

	l.log(LogTrace, "Unmounted", logEx{"name": l.Name})

	*l.updatesChannel <- ComponentLifeTimeMessage{
		Stage:     Unmounted,
		Component: l,
	}

	l.updatesChannel = nil

	return nil
}
