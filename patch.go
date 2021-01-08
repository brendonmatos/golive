package golive

type PatchInstruction struct {
	Name     string      `json:"n"`
	Type     string      `json:"t"`
	Attr     interface{} `json:"a,omitempty"`
	Content  string      `json:"c,omitempty"`
	Selector string      `json:"s"`
}

type PatchNodeChildren map[int]*PatchTreeNode

type PatchTreeNode struct {
	Children    PatchNodeChildren  `json:"c,omitempty"`
	Instruction []PatchInstruction `json:"i"`
}

type PatchBrowser struct {
	ComponentID  string             `json:"cid"`
	Name         string             `json:"n"`
	Instructions []PatchInstruction `json:"i"`
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
