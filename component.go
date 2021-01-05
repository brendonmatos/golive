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

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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
	Name       string
	IsMounted  bool
	IsPrepared bool
	IsCreated  bool
	Exited     bool

	log            Log
	component      ComponentLifeTime
	updatesChannel *ComponentLifeCycle
	htmlTemplate   *template.Template
	html           *html.Node
	rootNode       *html.Node
	rendered       string
}

// NewLiveComponent ...
func NewLiveComponent(name string, time ComponentLifeTime) *LiveComponent {
	return &LiveComponent{
		Name:      name,
		component: time,
		rootNode: &html.Node{
			Type:     html.ElementNode,
			Data:     "div",
			DataAtom: atom.Div,
		},
	}
}

func (l *LiveComponent) createUniqueName() string {
	return l.Name + "_" + NewLiveId().GenerateSmall()
}

func (l *LiveComponent) RenderChild(fn reflect.Value, _ ...reflect.Value) template.HTML {

	child, ok := fn.Interface().(*LiveComponent)

	if !ok {
		l.log(LogError, "child not a *golive.LiveComponent", nil)
		return ""
	}
	l.log(LogDebug, "will mount child", nil)

	l.log(LogDebug, "child mounted", nil)

	render, err := child.Render()
	if err != nil {
		l.log(LogError, "render child: render", logEx{"error": err})
	}

	return template.HTML(render)
}

func (l *LiveComponent) Create() error {
	var err error

	l.Name = l.createUniqueName()

	templateString := l.component.TemplateHandler(l)

	l.html, err = CreateDOMFromString(templateString)

	if err != nil {
		return err
	}

	l.rootNode.AppendChild(l.html)

	l.addWSConnectScript()
	l.addGoLiveComponentIDAttribute()
	l.addGoLiveElementsUID()

	templateString, err = RenderNodeToString(l.rootNode)

	if err != nil {
		return err
	}

	l.htmlTemplate, err = template.New(l.Name).Funcs(template.FuncMap{
		"render": l.RenderChild,
	}).Parse(templateString)

	l.CreateChildren()

	l.IsCreated = true

	return err
}

// Prepare 1.
func (l *LiveComponent) Prepare(updatesChannel *ComponentLifeCycle) error {

	l.updatesChannel = updatesChannel
	l.component.Prepare(l)
	l.PrepareChildren()

	l.IsPrepared = true

	return nil
}

func (l *LiveComponent) getChildrenComponents() []*LiveComponent {
	components := make([]*LiveComponent, 0)
	v := reflect.ValueOf(l.component).Elem()
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanInterface() {
			continue
		}

		lc, ok := v.Field(i).Interface().(*LiveComponent)
		if !ok {
			continue
		}

		components = append(components, lc)
	}
	return components
}

func (l *LiveComponent) CreateChildren() {
	for _, child := range l.getChildrenComponents() {
		child.Create()
	}
}

func (l *LiveComponent) MountChildren() {
	l.notifyStage(WillMountChildren)
	for _, child := range l.getChildrenComponents() {
		child.Mount()
	}
	l.notifyStage(ChildrenMounted)
}

func (l *LiveComponent) PrepareChildren() {
	l.notifyStage(WillPrepareChildren)
	for _, child := range l.getChildrenComponents() {
		child.Prepare(l.updatesChannel)
	}
	l.notifyStage(ChildrenPrepared)
}

// Mount 2. the component loading html
func (l *LiveComponent) Mount() error {

	if l.updatesChannel == nil {
		return fmt.Errorf("component is not prepared")
	}

	l.notifyStage(WillMount)

	l.component.BeforeMount(l)
	l.IsMounted = true

	l.component.Mounted(l)

	l.MountChildren()

	l.notifyStage(Mounted)

	return nil

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
func (l *LiveComponent) InvokeMethodInPath(path string, valuePath string) error {
	c := (*l).component
	v := reflect.ValueOf(c)

	var params []reflect.Value

	if len(valuePath) > 0 {
		params = append(params, *l.GetFieldFromPath(path))
	}

	v.MethodByName(path).Call(params)

	return nil
}

// Render ...
func (l *LiveComponent) Render() (string, error) {
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
func (l *LiveComponent) LiveRender() (*PatchBrowser, error) {
	newRender, err := l.Render()

	if err != nil {
		return nil, err
	}

	om := NewPatchBrowser(l.Name)

	om.Name = EventLiveDom

	if len(l.rendered) > 0 {

		changeInstructions, err := GetDiffFromRawHTML(l.rendered, newRender)

		if err != nil {
			l.log(LogPanic, "there is a error in diff", logEx{"error": err})
		}

		for _, instruction := range changeInstructions {

			selector, err := SelectorFromNode(instruction.Element)

			if err != nil {
				continue
				// l.log(LogPanic, "there is a error in selector", logEx{"error": err})
			}

			om.AddInstruction(PatchInstruction{
				Name:     EventLiveDom,
				Type:     strconv.Itoa(int(instruction.Type)),
				Attr:     instruction.Attr,
				Content:  instruction.Content,
				Selector: selector,
			})
		}
	}

	l.rendered = newRender

	return om, nil
}

var re = regexp.MustCompile(`<([a-z0-9]+)`)

func (l *LiveComponent) notifyStage(ltu LifeTimeStage) {
	*l.updatesChannel <- ComponentLifeTimeMessage{
		Stage:     ltu,
		Component: l,
	}
}

func (l *LiveComponent) addWSConnectScript() {
	l.rootNode.AppendChild(&html.Node{
		Type:      html.ElementNode,
		DataAtom:  atom.Script,
		Data:      "script",
		Namespace: "",
		FirstChild: &html.Node{
			Type: html.TextNode,
			Data: `goLive.once('WS_CONNECTION_OPEN', function() {
				goLive.connect('` + l.Name + `')
			})`,
		},
		Attr: []html.Attribute{
			{
				Key: "type",
				Val: "application/javascript",
			},
		},
	})
}

// TODO: improve this urgently
func (l *LiveComponent) addGoLiveComponentIDAttribute() {

	l.html.Attr = append(l.html.Attr, html.Attribute{
		Key: "go-live-component-id",
		Val: l.Name,
	})

}

func (l *LiveComponent) addGoLiveElementsUID() {
	// parent.FirstChild
	for index, node := range append(GetAllChildrenRecursive(l.html), l.html) {
		node.Attr = append(node.Attr, html.Attribute{
			Key: "go-live-uid",
			Val: l.Name + "-" + strconv.FormatInt(int64(index), 16),
		})
	}

}

// Kill ...
func (l *LiveComponent) Kill() error {

	*l.updatesChannel <- ComponentLifeTimeMessage{
		Stage:     WillUnmount,
		Component: l,
	}

	l.Exited = true
	// Set all to nil to garbage collector act
	l.component = nil
	l.updatesChannel = nil
	l.htmlTemplate = nil

	// *l.updatesChannel <- ComponentLifeTimeMessage{
	// 	Stage:     Unmounted,
	// 	Component: l,
	// }

	return nil
}
