package golive

import (
	"testing"
)

type Pet struct {
	LiveComponentWrapper
	Name  string
	Age   int
	Awake bool
}

var petComponent = NewLiveComponent("pet", &Pet{
	Name: "Catdog",
	Age:  12,
})

func TestLiveComponent_GetFieldFromPath(t *testing.T) {
	field := petComponent.GetFieldFromPath("Name")

	if field.String() != "Catdog" {
		t.Error("The get field should return Catdog")
	}
}

func TestLiveComponent_SetValueInPathWithString(t *testing.T) {
	err := petComponent.SetValueInPath("Dog", "Name")

	if err != nil {
		t.Error(err)
		return
	}

	field := petComponent.GetFieldFromPath("Name")

	if field.String() == "Catdog" {
		t.Error("The field has not been set")
	}

	if field.String() != "Dog" {
		t.Errorf("The field has set with different value! with ->%v", field)
	}
}

func TestLiveComponent_SetValueInPathWithNumber(t *testing.T) {
	err := petComponent.SetValueInPath("10", "Age")

	if err != nil {
		t.Error(err)
		return
	}

	field := petComponent.GetFieldFromPath("Age")

	if field.Int() == 12 {
		t.Error("The field has not been set")
	}

	if field.Int() != 10 {
		t.Error("The field has not been set")
	}
}

func TestLiveComponent_SetValueInPathWithBoolean(t *testing.T) {
	err := petComponent.SetValueInPath("true", "Awake")

	if err != nil {
		t.Error(err)
		return
	}

	field := petComponent.GetFieldFromPath("Awake")

	if !field.Bool() {
		t.Error("The field has not been set")
	}
}

func TestLiveComponent_SetValueInPathWithBoolean2(t *testing.T) {
	err := petComponent.SetValueInPath("false", "Awake")

	if err != nil {
		t.Error(err)
		return
	}

	field := petComponent.GetFieldFromPath("Awake")

	if field.Bool() {
		t.Error("The field has not been set")
	}
}
