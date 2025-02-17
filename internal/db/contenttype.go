package db

type ContentEnumerator interface {
	String() string
	Index() int
}

const (
	ContentText ContentType = iota
	ContentPhoto
	ContentVideo
)

type ContentType int

func (ct ContentType) String() string {
	return [...]string{"TEXT", "PHOTO", "VIDEO"}[ct]
}

func (ct ContentType) Index() int {
	return int(ct)
}

type Content struct {
	Type ContentEnumerator `bson:"type"`
	Path string            `bson:"path"`
	Text string            `bson:"text"`
	File []byte            `bson:"-"`
}
