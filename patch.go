package golive

type PatchInstruction struct {
	Name    string      `json:"n"`
	Type    string      `json:"t"`
	Attr    interface{} `json:"a,omitempty"`
	Content string      `json:"c,omitempty"`
}

type PatchNodeChildren map[int]*PatchTreeNode

type PatchTreeNode struct {
	Children    PatchNodeChildren  `json:"c,omitempty"`
	Instruction []PatchInstruction `json:"i,omitempty"`
}

func NewPatchTreeNode() *PatchTreeNode {
	return &PatchTreeNode{
		Children:    make(PatchNodeChildren),
		Instruction: make([]PatchInstruction, 0),
	}
}

type PatchBrowser struct {
	ComponentID string         `json:"i"`
	Name        string         `json:"n"`
	Root        *PatchTreeNode `json:"r"`
}

func NewPatchBrowser(componentId string) *PatchBrowser {
	return &PatchBrowser{
		ComponentID: componentId,
		Root:        NewPatchTreeNode(),
	}
}

func (pt *PatchTreeNode) AddPathInstruction(path []int, pi PatchInstruction) {
	if len(path) == 0 {
		pt.Instruction = append(pt.Instruction, pi)
		return
	}

	if pt.Children == nil {
		pt.Children = make(PatchNodeChildren)
	}

	index := path[0]

	child, ok := pt.Children[index]

	if !ok {
		child = NewPatchTreeNode()
		pt.Children[index] = child
	}

	child.AddPathInstruction(path[1:], pi)
}
