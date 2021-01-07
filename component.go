package golive

import (
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

//
type ComponentLifeTime interface {
	Create(component *LiveComponent)
	TemplateHandler(component *LiveComponent) string
	Mounted(component *LiveComponent)
	BeforeMount(component *LiveComponent)
}

type ChildLiveComponent interface{}

//
type LiveComponent struct {
	Name string

	IsMounted bool
	IsCreated bool
	Exited    bool

	log       Log
	life      *ComponentLifeCycle
	component ComponentLifeTime
	renderer  LiveRenderer

	children []*LiveComponent
}

// NewLiveComponent ...
func NewLiveComponent(name string, time ComponentLifeTime) *LiveComponent {
	return &LiveComponent{
		Name:      name,
		component: time,
		renderer: LiveRenderer{
			state:      &LiveState{},
			template:   nil,
			formatters: make([]func(t string) string, 0),
		},
	}
}

func (l *LiveComponent) Create(life *ComponentLifeCycle) error {
	var err error

	l.life = life

	// The first notification, will notify
	// an component without unique name
	l.notifyStage(WillCreate)

	l.Name = l.createUniqueName()

	// Get the template defined on component
	ts := l.component.TemplateHandler(l)

	// Prepare the template content adding
	// golive specific
	ts = l.addGoLiveComponentIDAttribute(ts)

	// Generate go std template
	ct, err := l.generateTemplate(ts)

	if err != nil {
		return err
	}

	l.renderer.setTemplate(ct, ts)

	//
	l.renderer.useFormatter(func(t string) string {
		d, _ := CreateDOMFromString(t)
		l.signRender(d)
		t, _ = RenderNodeChildren(d)
		return t
	})

	// Calling component creation
	l.component.Create(l)

	// Creating children
	err = l.createChildren()

	if err != nil {
		return err
	}

	l.IsCreated = true

	l.notifyStage(Created)

	return err
}

func (l *LiveComponent) createChildren() error {
	var err error
	for _, child := range l.getChildrenComponents() {
		err = child.Create(l.life)
		if err != nil {
			panic(err)
		}

		l.children = append(l.children, child)
	}
	return err
}

func (l *LiveComponent) findComponentById(id string) *LiveComponent {
	if l.Name == id {
		return l
	}

	for _, child := range l.children {
		if child.Name == id {
			return l
		}
	}

	for _, child := range l.children {
		found := child.findComponentById(id)

		if found != nil {
			return found
		}
	}

	return nil
}

// Mount 2. the component loading html
func (l *LiveComponent) Mount() error {

	if !l.IsCreated {
		return fmt.Errorf("component need to be prepared")
	}

	l.notifyStage(WillMount)

	l.component.BeforeMount(l)
	l.component.Mounted(l)
	l.MountChildren()
	l.IsMounted = true

	l.notifyStage(Mounted)

	return nil
}

func (l *LiveComponent) MountChildren() {
	l.notifyStage(WillMountChildren)
	for _, child := range l.getChildrenComponents() {
		_ = child.Mount()
	}
	l.notifyStage(ChildrenMounted)
}

// Render ...
func (l *LiveComponent) Render() (string, error) {
	text, _, err := l.renderer.Render(l.component)
	return text, err
}

func (l *LiveComponent) RenderChild(fn reflect.Value, _ ...reflect.Value) template.HTML {

	child, ok := fn.Interface().(*LiveComponent)

	if !ok {
		l.log(LogError, "child not a *golive.LiveComponent", nil)
		return ""
	}

	render, err := child.Render()
	if err != nil {
		l.log(LogError, "render child: render", logEx{"error": err})
	}

	return template.HTML(render)
}

// LiveRender render a new version of the component, and detect
// differences from the last render
// and sets the "new old" version  of render
func (l *LiveComponent) LiveRender() (*Diff, error) {
	return l.renderer.LiveRender(l.component)
}

func (l *LiveComponent) Update() {
	l.notifyStage(Updated)
}

// Kill ...
func (l *LiveComponent) Kill() error {

	l.notifyStage(WillUnmount)

	l.Exited = true
	l.component = nil
	l.life = nil

	l.notifyStage(Unmounted)

	return nil
}

// GetFieldFromPath ...
func (l *LiveComponent) GetFieldFromPath(path string) *reflect.Value {
	c := (*l).component
	v := reflect.ValueOf(c).Elem()

	for _, s := range strings.Split(path, ".") {

		if reflect.ValueOf(v).IsZero() {
			l.log(LogError, "field not found in component", logEx{
				"component": l.Name,
				"path":      path,
			})
		}

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

func (l *LiveComponent) createUniqueName() string {
	return l.Name + "_" + NewLiveId().GenerateSmall()
}

// TODO: maybe nested components?
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

func (l *LiveComponent) notifyStage(ltu LifeTimeStage) {
	*l.life <- ComponentLifeTimeMessage{
		Stage:     ltu,
		Component: l,
	}
}

var rxTagName = regexp.MustCompile(`<([a-z0-9]+)`)

func (l *LiveComponent) addGoLiveComponentIDAttribute(template string) string {
	found := rxTagName.FindString(template)
	if found != "" {
		replaceWith := found + ` go-live-component-id="` + l.Name + `" `
		template = strings.Replace(template, found, replaceWith, 1)
	}
	return template
}

func (l *LiveComponent) generateTemplate(ts string) (*template.Template, error) {
	return template.New(l.Name).Funcs(template.FuncMap{
		"render": l.RenderChild,
	}).Parse(ts)
}

func (l *LiveComponent) signRender(dom *html.Node) error {

	// Post treatment
	for index, node := range GetAllChildrenRecursive(dom) {

		if goLiveIdAttr := getAttribute(node, "go-live-uid"); goLiveIdAttr == nil {
			addNodeAttribute(node, "go-live-uid", strconv.FormatInt(int64(index), 16))
		}

		if goLiveInputAttr := getAttribute(node, "go-live-input"); goLiveInputAttr != nil {
			addNodeAttribute(node, ":value", goLiveInputAttr.Val)
		}

		if valueAttr := getAttribute(node, ":value"); valueAttr != nil {
			removeNodeAttribute(node, ":value")

			cid, err := ComponentIdFromNode(node)

			if err != nil {
				return err
			}

			c := l.findComponentById(cid)

			if c == nil {
				return fmt.Errorf("component not found")
			}

			f := c.GetFieldFromPath(valueAttr.Val)

			if inputTypeAttr := getAttribute(node, "type"); inputTypeAttr != nil {
				switch inputTypeAttr.Val {
				case "checkbox":
					if f.Bool() {
						addNodeAttribute(node, "checked", "checked")
					} else {
						removeNodeAttribute(node, "checked")
					}
					break
				}
			} else {
				addNodeAttribute(node, "value", fmt.Sprintf("%v", f))
			}
		}

		if disabledAttr := getAttribute(node, ":disabled"); disabledAttr != nil {
			removeNodeAttribute(node, ":disabled")
			if disabledAttr.Val == "true" {
				addNodeAttribute(node, "disabled", "disabled")
			} else {
				removeNodeAttribute(node, "disabled")
			}
		}

	}

	return nil
}

func ComponentIdFromNode(e *html.Node) (string, error) {
	for parent := e; parent != nil; parent = parent.Parent {
		if componentAttr := getAttribute(parent, "go-live-component-id"); componentAttr != nil {
			return componentAttr.Val, nil
		}
	}
	return "", fmt.Errorf("node not found")
}
