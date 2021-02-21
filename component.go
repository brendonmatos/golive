package golive

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html/atom"

	"golang.org/x/net/html"
)

const ComponentIdAttrKey = "go-live-component-id"

var (
	ErrComponentNotPrepared = errors.New("Component need to be prepared")
	ErrComponentWithoutLog  = errors.New("Component without log defined")
	ErrComponentNil         = errors.New("Component nil")
)

//
type ComponentLifeTime interface {
	Create(component *LiveComponent)
	TemplateHandler(component *LiveComponent) string
	Mounted(component *LiveComponent)
	BeforeMount(component *LiveComponent)
	BeforeUnmount(component *LiveComponent)
}

type ChildLiveComponent interface{}

type ComponentContext struct {
	Pairs map[string]interface{}
}

func NewComponentContext() ComponentContext {
	return ComponentContext{
		Pairs: map[string]interface{}{},
	}
}

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

	Context ComponentContext
}

// NewLiveComponent ...
func NewLiveComponent(name string, component ComponentLifeTime) *LiveComponent {
	return &LiveComponent{
		Name:      name,
		component: component,
		Context:   NewComponentContext(),
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

	if l.log == nil {
		return ErrComponentWithoutLog
	}

	// The first notification, will notify
	// an Component without unique name
	l.notifyStage(WillCreate)

	l.Name = l.createUniqueName()

	// Get the template defined on Component
	ts := l.component.TemplateHandler(l)

	// Prepare the template content adding
	// golive specific
	ts = l.addGoLiveComponentIDAttribute(ts)
	ts = l.signTemplateString(ts)

	// Generate go std template
	ct, err := l.generateTemplate(ts)

	if err != nil {
		return fmt.Errorf("generate template: %w", err)
	}

	l.renderer.setTemplate(ct, ts)

	//
	l.renderer.useFormatter(func(t string) string {
		d, _ := nodeFromString(t)
		_ = l.treatRender(d)
		t, _ = renderInnerHTML(d)
		return t
	})

	// Calling Component creation
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
		child.log = l.log
		child.Context = l.Context
		err = child.Create(l.life)
		if err != nil {
			panic(err)
		}

		l.children = append(l.children, child)
	}
	return err
}

func (l *LiveComponent) findComponentByID(id string) *LiveComponent {
	if l.Name == id {
		return l
	}

	for _, child := range l.children {
		if child.Name == id {
			return child
		}
	}

	for _, child := range l.children {
		found := child.findComponentByID(id)

		if found != nil {
			return found
		}
	}

	return nil
}

// Mount 2. the Component loading html
func (l *LiveComponent) Mount() error {

	if !l.IsCreated {
		return ErrComponentNotPrepared
	}

	l.notifyStage(WillMount)

	l.component.BeforeMount(l)

	err := l.MountChildren()

	if err != nil {
		return fmt.Errorf("mount children: %w", err)
	}

	l.component.Mounted(l)

	l.IsMounted = true

	l.notifyStage(Mounted)

	return nil
}

func (l *LiveComponent) MountChildren() error {
	l.notifyStage(WillMountChildren)
	for _, child := range l.getChildrenComponents() {
		err := child.Mount()

		if err != nil {
			return fmt.Errorf("child mount: %w", err)
		}
	}
	l.notifyStage(ChildrenMounted)
	return nil
}

// Render ...
func (l *LiveComponent) Render() (string, error) {
	l.log(LogTrace, "Render", logEx{"name": l.Name})

	if l.component == nil {
		return "", ErrComponentNil
	}

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

// LiveRender render a new version of the Component, and detect
// differences from the last render
// and sets the "new old" version  of render
func (l *LiveComponent) LiveRender() (*diff, error) {
	return l.renderer.LiveRender(l.component)
}

func (l *LiveComponent) Update() {
	l.notifyStage(Updated)
}

func (l *LiveComponent) UpdateWithSource(source *EventSource) {
	l.notifyStageWithSource(Updated, source)
}

// Kill ...
func (l *LiveComponent) Kill() error {

	l.KillChildren()

	l.log(LogTrace, "WillUnmount", logEx{"name": l.Name})

	l.component.BeforeUnmount(l)

	l.notifyStage(WillUnmount)

	l.Exited = true
	l.component = nil

	l.notifyStage(Unmounted)

	l.life = nil

	return nil
}

func (l *LiveComponent) KillChildren() {
	for _, child := range l.children {
		if err := child.Kill(); err != nil {
			l.log(LogError, "kill child", logEx{"name": child.Name})
		}
	}
}

// GetFieldFromPath ...
func (l *LiveComponent) GetFieldFromPath(path string) *reflect.Value {
	c := (*l).component
	v := reflect.ValueOf(c).Elem()

	for _, s := range strings.Split(path, ".") {

		if reflect.ValueOf(v).IsZero() {
			l.log(LogError, "field not found in Component", logEx{
				"Component": l.Name,
				"path":      path,
			})
		}

		if v.Kind() == reflect.Ptr {
			v = v.Elem()
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

func jsonEscape(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	s := string(b)
	return s[1 : len(s)-1]
}

// SetValueInPath ...
func (l *LiveComponent) SetValueInPath(value string, path string) error {

	v := l.GetFieldFromPath(path)
	n := reflect.New(v.Type())

	if v.Kind() == reflect.String {
		value = `"` + jsonEscape(value) + `"`
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

func (l *LiveComponent) createUniqueName() string {
	return l.Name + "_" + NewLiveID().GenerateSmall()
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

func (l *LiveComponent) notifyStage(ltu LifeTimeStage) {
	l.notifyStageWithSource(ltu, nil)
}

func (l *LiveComponent) notifyStageWithSource(ltu LifeTimeStage, source *EventSource) {
	if l.life == nil {
		l.log(LogWarn, "Component life updates channel is nil", nil)
		return
	}

	*l.life <- ComponentLifeTimeMessage{
		Stage:     ltu,
		Component: l,
		Source:    source,
	}
}

var rxTagName = regexp.MustCompile(`<([a-z0-9]+[ ]?)`)

func (l *LiveComponent) addGoLiveComponentIDAttribute(template string) string {
	found := rxTagName.FindString(template)
	if found != "" {
		replaceWith := found + ` ` + ComponentIdAttrKey + `="` + l.Name + `" `
		template = strings.Replace(template, found, replaceWith, 1)
	}
	return template
}

func (l *LiveComponent) generateTemplate(ts string) (*template.Template, error) {
	return template.New(l.Name).Funcs(template.FuncMap{
		"render": l.RenderChild,
	}).Parse(ts)
}

func (l *LiveComponent) treatRender(dom *html.Node) error {

	// Post treatment
	for _, node := range getAllChildrenRecursive(dom) {

		if goLiveInputAttr := getAttribute(node, "go-live-input"); goLiveInputAttr != nil {
			addNodeAttribute(node, ":value", goLiveInputAttr.Val)
		}

		if valueAttr := getAttribute(node, ":value"); valueAttr != nil {
			removeNodeAttribute(node, ":value")

			cid, err := componentIDFromNode(node)

			if err != nil {
				return err
			}

			foundComponent := l.findComponentByID(cid)

			if foundComponent == nil {
				return fmt.Errorf("Component not found")
			}

			f := foundComponent.GetFieldFromPath(valueAttr.Val)

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
			} else if node.DataAtom == atom.Textarea {
				n, err := nodeFromString(fmt.Sprintf("%v", f))

				if n == nil || n.FirstChild == nil {
					continue
				}

				if err != nil {
					continue
				}

				child := n.FirstChild

				n.RemoveChild(child)

				node.AppendChild(child)
			} else {
				addNodeAttribute(node, "value", fmt.Sprintf("%v", f))
			}
		}

		if disabledAttr := getAttribute(node, ":disabled"); disabledAttr != nil {
			removeNodeAttribute(node, ":disabled")
			if disabledAttr.Val == "true" {
				addNodeAttribute(node, "disabled", "")
			} else {
				removeNodeAttribute(node, "disabled")
			}
		}
	}
	return nil
}

func (l *LiveComponent) signTemplateString(ts string) string {
	matches := rxTagName.FindAllStringSubmatchIndex(ts, -1)

	reverseSlice(matches)

	for _, match := range matches {
		startIndex := match[0]
		endIndex := match[1]

		startSlice := ts[:startIndex]
		endSlide := ts[endIndex:]
		matchedSlice := ts[startIndex:endIndex]

		uid := l.Name + "_" + NewLiveID().GenerateSmall()
		replaceWith := matchedSlice + ` go-live-uid="` + uid + `" `
		ts = startSlice + replaceWith + endSlide
	}

	return ts
}

func componentIDFromNode(e *html.Node) (string, error) {
	for parent := e; parent != nil; parent = parent.Parent {
		if componentAttr := getAttribute(parent, ComponentIdAttrKey); componentAttr != nil {
			return componentAttr.Val, nil
		}
	}
	return "", fmt.Errorf("node not found")
}

func reverseSlice(s interface{}) {
	size := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, size-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
