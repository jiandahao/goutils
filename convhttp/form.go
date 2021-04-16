package convhttp

import (
	"encoding/json"
	"net/url"
)

// FormData form data
type FormData struct {
	url.Values
	file *formFile
}

// NewFormData new from data instance
func NewFormData() *FormData {
	return &FormData{
		Values: url.Values{},
	}
}

type formFile struct {
	fieldname string
	filename  string
	data      []byte
}

func (ff *formFile) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		FieldName string `json:"field_name"`
		Filename  string `json:"file_name"`
	}{
		FieldName: ff.fieldname,
		Filename:  ff.filename,
	})
}

// WithFile with file
func (fd *FormData) WithFile(fieldname string, filename string, data []byte) {
	fd.file = &formFile{
		fieldname: fieldname,
		filename:  filename,
		data:      data,
	}
}

// MarshalJSON implements MarshalJSON method
// to produce JSON.
func (fd *FormData) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Fileds url.Values `json:"fileds"`
		File   *formFile  `json:"file,omitempty"`
	}{
		Fileds: fd.Values,
		File:   fd.file,
	})
}
