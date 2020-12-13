package golive

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"reflect"
)

type InMessage struct {
	Name         string `json:"name"`
	ScopeID      string `json:"scope_id"`
	MethodName   string `json:"method_name"`
	MethodParams string `json:"method_params"`
	StateKey     string `json:"key"`
	StateValue   string `json:"value"`
}

type OutMessage struct {
	Name    string      `json:"name"`
	ScopeID string      `json:"scope_id"`
	Type    string      `json:"type"`
	Attr    interface{} `json:"attr"`
	Content string      `json:"content"`
	Element string      `json:"element"`
}

const (
	EventLiveInput      = "li"
	EventLiveMethod     = "lm"
	EventLiveDom        = "ld"
	EventLiveDisconnect = "lx"
)

var BasePage *template.Template
var LiveLib string

func init() {
	basePageBytes, _ := ioutil.ReadFile("./assets/base_page.html")
	basePage := string(basePageBytes)
	BasePage, _ = template.New("BasePage").Parse(basePage)
}

type LivePage struct {
	Session                   SessionKey
	Component                 *LiveComponent
	ComponentsLifeTimeChannel *LiveTimeChannel
	RenderedContent           string
}

type PageContent struct {
	Lang   string
	Body   string `html:"unsafe"`
	Script string `html:"unsafe"`
	Title  string
	Enum   PageEnum
}

type PageEnum struct {
	EventLiveInput  string
	EventLiveMethod string
	EventLiveDom    string
}

type LivePageInterface interface {
	HandleMessage(m InMessage)
}

func NewLivePageToComponent(s SessionKey, c *LiveComponent) *LivePage {
	channel := make(LiveTimeChannel)

	c.LifeTimeChannel = &channel

	return &LivePage{
		Session:                   s,
		Component:                 c,
		ComponentsLifeTimeChannel: &channel,
	}
}

func (lp *LivePage) Mount() {
	lp.Component.Prepare()
	lp.Component.Mount(lp.ComponentsLifeTimeChannel)
}

func asUnsafeMap(any interface{}) map[string]interface{} {
	v := reflect.ValueOf(any)

	m := map[string]interface{}{}
	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i)
		if !value.CanInterface() {
			continue
		}
		ftype := v.Type().Field(i)
		if ftype.Tag.Get("html") == "unsafe" {
			m[ftype.Name] = template.HTML(value.String())
		} else {
			m[ftype.Name] = value.Interface()
		}
	}
	return m
}

func (lp *LivePage) FirstRender(pc PageContent) string {
	rendered := lp.Component.GetComponentRender()
	writer := bytes.NewBufferString("")

	pc.Body = rendered
	pc.Script = LiveLib
	pc.Enum = PageEnum{
		EventLiveDom:    EventLiveDom,
		EventLiveInput:  EventLiveInput,
		EventLiveMethod: EventLiveMethod,
	}

	_ = BasePage.Execute(writer, asUnsafeMap(pc))

	return writer.String()
}
