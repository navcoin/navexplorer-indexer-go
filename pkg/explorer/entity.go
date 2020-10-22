package explorer

type Entity interface {
	Id() string
	SetId(id string)
	Slug() string
}
