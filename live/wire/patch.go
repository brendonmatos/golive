package wire

type PatchInstruction struct {
	Name     string      `json:"n"`
	Type     string      `json:"t"`
	Attr     interface{} `json:"a,omitempty"`
	Content  string      `json:"c,omitempty"`
	Selector string      `json:"s"`
	Index    int         `json:"i,omitempty"`
}

type PatchBrowser struct {
	ComponentID  string             `json:"cid,omitempty"`
	Type         string             `json:"t"`
	Message      string             `json:"m"`
	Instructions []PatchInstruction `json:"i,omitempty"`
	Value        interface{}        `json:"value"`
}

func NewPatchBrowser(componentID string) *PatchBrowser {
	return &PatchBrowser{
		ComponentID:  componentID,
		Instructions: make([]PatchInstruction, 0),
	}
}

func (pb *PatchBrowser) AddInstruction(pi PatchInstruction) {
	pb.Instructions = append(pb.Instructions, pi)
}
