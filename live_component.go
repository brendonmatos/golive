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

type UpdateMessage int

type LiveTimeChannel chan UpdateMessage

const (
	LifeTimeExit = iota
	LifeTimeUpdate
)

//
type ComponentLifeTime interface {
	TemplateHandler() string
	Mounted(component *LiveComponent)
	BeforeMount(component *LiveComponent)
	SetLifeTimeChannel(c *LiveTimeChannel)
}

//
type LiveComponent struct {
	Name               string
	Component          ComponentLifeTime
	LifeTimeChannel    *LiveTimeChannel
	HTMLTemplateString string
	HTMLTemplate       *template.Template
	Rendered           string
	IsMounted          bool
}

func (l *LiveComponent) getName() string {
	return l.Name + "_" + NewLiveId().GenerateRandomString()
}

func NewLiveComponent(name string, time ComponentLifeTime) *LiveComponent {
	return &LiveComponent{
		Name:      name,
		Component: time,
	}
}

//  Prepare 1.
func (l *LiveComponent) Prepare() {
	l.Name = l.getName()
	l.HTMLTemplateString = l.Component.TemplateHandler()

	l.HTMLTemplateString = l.addWSConnectScript(l.HTMLTemplateString)
	l.HTMLTemplateString = l.addGoLiveComponentIdAttribute(l.HTMLTemplateString)
	l.HTMLTemplate, _ = template.New(l.Name).Parse(l.HTMLTemplateString)
	l.Component.SetLifeTimeChannel(l.LifeTimeChannel)
}

// Mount 2. the component loading html
func (l *LiveComponent) Mount(a *LiveTimeChannel) {
	l.Component.BeforeMount(l)
	l.IsMounted = true
	l.LifeTimeChannel = a
	l.Component.Mounted(l)

}

func (l *LiveComponent) FindComponent(_ string) (*LiveComponent, error) {
	return l, nil
}

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

func (l *LiveComponent) SetValueInPath(value string, path string) error {
	v := l.GetFieldFromPath(path)
	n := reflect.New(v.Type())

	if v.Kind() == reflect.String {
		value = fmt.Sprintf("\"%s\"", value)
	}

	err := json.Unmarshal([]byte(value), n.Interface())
	if err != nil {
		fmt.Println(value, err)
		return err
	}

	v.Set(n.Elem())
	return nil
}

func (l *LiveComponent) InvokeMethodInPath(path string, valuePath string) {
	c := (*l).Component
	v := reflect.ValueOf(c)

	var params []reflect.Value

	if len(valuePath) > 0 {
		params = append(params, *l.GetFieldFromPath(path))
	}

	v.MethodByName(path).Call(params)
}

func (l *LiveComponent) GetComponentRender() string {
	s := bytes.NewBufferString("")
	_ = l.HTMLTemplate.Execute(s, l.Component)
	return s.String()
}

// LiveRender render the last version of the component, and detect
// differences from the last render
// and sets the new version of render
func (l *LiveComponent) LiveRender() (string, []OutMessage) {
	newRender := l.GetComponentRender()

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

func (l *LiveComponent) addGoLiveComponentIdAttribute(template string) string {
	found := re.FindString(l.HTMLTemplateString)
	if found != "" {
		replaceWith := found + ` go-live-component-id="` + l.Name + `" `
		template = strings.Replace(l.HTMLTemplateString, found, replaceWith, 1)
	}
	return template
}

func (l *LiveComponent) Kill() error {
	*l.LifeTimeChannel <- LifeTimeExit
	l.Component = nil
	l.LifeTimeChannel = nil
	l.HTMLTemplate = nil
	return nil
}
