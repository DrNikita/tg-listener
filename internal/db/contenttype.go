package db

type ContentEnumerator interface {
	String() string
	Index() int
}

const (
	ContentText ContentType = iota
	ContentPhoto
	ContentVideo
	ContentVoice
	ContentDocument

	ContentTextString     string = "TEXT"
	ContentPhotoString    string = "PHOTO"
	ContentVideoString    string = "VIDEO"
	ContentVoiceString    string = "VOICE"
	ContentDocumentString string = "DOCUMENT"
)

type ContentType int

func (ct ContentType) String() string {
	return [...]string{ContentTextString, ContentPhotoString, ContentVideoString}[ct]
}

func (ct ContentType) Index() int {
	return int(ct)
}

type Content struct {
	Type    ContentEnumerator `bson:"type"`
	Path    string            `bson:"path"`
	Text    string            `bson:"text"`
	MediaID int32             `bson:"-"`
}
