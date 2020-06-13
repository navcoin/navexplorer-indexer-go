package explorer

import "github.com/gosimple/slug"

type Cfund struct {
	Available float64 `json:"available"`
	Locked    float64 `json:"locked"`
}

func (c *Cfund) Slug() string {
	return slug.Make("cfund")
}
