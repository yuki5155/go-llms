package consts

type Model string

const GPT4o20240806 Model = "gpt-4o-2024-08-06"
const DefaultModel Model = GPT4o20240806

func NewDefaultModel() Model {
	return GPT4o20240806
}

func (m Model) String() string {
	return string(m)
}
