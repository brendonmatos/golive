package live

import (
	"github.com/brendonmatos/golive"
	"sync"
	"testing"
	"time"
)

type Pet struct {
	Wrapper
	Name  string
	Age   int
	Awake bool
}

var petComponent = NewLiveComponent("pet", &Pet{
	Name: "Catdog",
	Age:  12,
})

func TestLiveComponent_GetFieldFromPath(t *testing.T) {
	field, _ := petComponent.State.GetFieldFromPath("Name")

	if field.String() != "Catdog" {
		t.Error("The get field should return Catdog")
	}
}

func TestLiveComponent_SetValueInPathWithString(t *testing.T) {
	err := petComponent.State.SetValueInPath("Dog", "Name")

	if err != nil {
		t.Error(err)
		return
	}

	field, _ := petComponent.State.GetFieldFromPath("Name")

	if field.String() == "Catdog" {
		t.Error("The field has not been set")
	}

	if field.String() != "Dog" {
		t.Errorf("The field has set with different value! with ->%v", field)
	}
}

func TestLiveComponent_SetValueInPathWithNumber(t *testing.T) {
	err := petComponent.State.SetValueInPath("10", "Age")

	if err != nil {
		t.Error(err)
		return
	}

	field, _ := petComponent.State.GetFieldFromPath("Age")

	if field.Int() == 12 {
		t.Error("The field has not been set")
	}

	if field.Int() != 10 {
		t.Error("The field has not been set")
	}
}

func TestLiveComponent_SetValueInPathWithBoolean(t *testing.T) {
	err := petComponent.State.SetValueInPath("true", "Awake")

	if err != nil {
		t.Error(err)
		return
	}

	field, _ := petComponent.State.GetFieldFromPath("Awake")

	if !field.Bool() {
		t.Error("The field has not been set")
	}
}

func TestLiveComponent_SetValueInPathWithBoolean2(t *testing.T) {
	err := petComponent.State.SetValueInPath("false", "Awake")

	if err != nil {
		t.Error(err)
		return
	}

	field, _ := petComponent.State.GetFieldFromPath("Awake")

	if field.Bool() {
		t.Error("The field has not been set")
	}
}

type Clock struct {
	Wrapper
}

func NewClock() *Component {
	return NewLiveComponent("Clock", &Clock{})
}

func (c *Clock) ActualTime() string {
	return time.Now().Format(time.RFC3339Nano)
}

func (c *Clock) Mounted(l *Component) {
	go func() {
		for {
			if l.Exited {
				return
			}
			time.Sleep(time.Second)
			c.Commit()
		}
	}()
}

func (c *Clock) TemplateHandler(_ *Component) string {
	return `
		<div>
			<span>Time: {{ .ActualTime }}</span>
		</div>
	`
}

func TestComponent_LifeCycleSequence(t *testing.T) {

	c := NewClock()

	c.Log = golive.NewLoggerBasic().Log

	lc := make(LifeCycle)

	desired := []LifeTimeStage{
		WillCreate,
		Created,
		WillMount,
		WillMountChildren,
		ChildrenMounted,
		Mounted,
		Rendered,
		Updated,
		WillUnmount,
		Unmounted,
	}

	wg := sync.WaitGroup{}

	// Test until mounted
	wg.Add(5)

	go func() {
		count := 0
		for {
			a := <-lc

			if desired[count] != a.Stage {
				t.Error("Stage not expected, expecting", desired[count], "received", a.Stage)
			}

			count++

			if a.Stage == Mounted {
				return
			}

			wg.Done()

		}
	}()

	err := c.Create()
	if err != nil {
		t.Error(err)
	}

	wg.Wait()
}

type TestComp struct {
	Wrapper
}

func (tc *TestComp) TemplateHandler(_ *Component) string {
	return `
		<div>
			<div></div>
			<div>
				<div></div>
			</div>
			<div></div>
			<div></div>
		</div>
	`
}

func TestComponent_ComponentSignTemplate(t *testing.T) {
	var err error
	c := NewLiveComponent("Test", &TestComp{})
	c.Log = golive.NewLoggerBasic().Log
	err = c.Create(nil)

	if err != nil {
		t.Error(err)
	}

	err = c.Mount()

	if err != nil {
		t.Error(err)
	}

}
