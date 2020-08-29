package image

type Transformer struct{}

func NewTransformer() *Transformer {
	return &Transformer{}
}

type TransformedLinks struct {
	Links []string `json:"links"`
}

func (t *Transformer) TransformLinks(links []string) *TransformedLinks {
	return &TransformedLinks{
		Links: links,
	}
}
