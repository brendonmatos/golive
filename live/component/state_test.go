package component

import "testing"

type Pet struct {
	Name  string
	Age   int
	Awake bool
}

var petComponent = &Pet{
	Name: "Catdog",
	Age:  12,
}

var state = NewState()

func init() {
	state.Set(petComponent)
}

func TestState_GetFieldFromPath(t *testing.T) {
	field, _ := state.GetFieldFromPath("Name")

	if field.String() != "Catdog" {
		t.Error("The get field should return Catdog")
	}
}

func TestState_SetValueInPathWithString(t *testing.T) {
	err := state.SetValueInPath("Dog", "Name")

	if err != nil {
		t.Error(err)
		return
	}

	field, _ := state.GetFieldFromPath("Name")

	if field.String() == "Catdog" {
		t.Error("The field has not been set")
	}

	if field.String() != "Dog" {
		t.Errorf("The field has set with different value! with ->%v", field)
	}
}

func TestState_SetValueInPathWithNumber(t *testing.T) {
	err := state.SetValueInPath("10", "Age")

	if err != nil {
		t.Error(err)
		return
	}

	field, _ := state.GetFieldFromPath("Age")

	if field.Int() == 12 {
		t.Error("The field has not been set")
	}

	if field.Int() != 10 {
		t.Error("The field has not been set")
	}
}

func TestState_SetValueInPathWithBoolean(t *testing.T) {
	err := state.SetValueInPath("true", "Awake")

	if err != nil {
		t.Error(err)
		return
	}

	field, _ := state.GetFieldFromPath("Awake")

	if !field.Bool() {
		t.Error("The field has not been set")
	}
}

func TestState_SetValueInPathWithBoolean2(t *testing.T) {
	err := state.SetValueInPath("false", "Awake")

	if err != nil {
		t.Error(err)
		return
	}

	field, _ := state.GetFieldFromPath("Awake")

	if field.Bool() {
		t.Error("The field has not been set")
	}
}
