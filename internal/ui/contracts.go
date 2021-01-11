package ui

// Document is the document
type Document struct {
	// ID is the document id
	ID string `json:id`
	// Name is the name
	Name string `json:name`
}

// DocumentList is the list of documents
type DocumentList struct {
	// Documents are the documents
	Documents []Document `json:documents`
}

type loginForm struct {
	Email    string `json:"Email"`
	Password string `json:"Password"`
}
