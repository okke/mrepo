package mrepo

import (
	"fmt"

	"github.com/jinzhu/inflection"
)

type Document interface {
	Data() map[string]interface{}
	Collection() string
	ID() string
	IDKey() string
}

type document struct {
	collection string
	data       map[string]interface{}
}

func D(collection string, dataSets ...map[string]interface{}) Document {

	newData := make(map[string]interface{}, 0)
	for _, data := range dataSets {
		for k, v := range data {

			newData[k] = v

		}
	}

	return &document{collection: collection, data: newData}
}

func (document *document) Data() map[string]interface{} {
	return document.data
}

func (document *document) IDKey() string {
	return fmt.Sprintf("%s_id", inflection.Singular(document.collection))
}

func (document *document) ID() string {

	objID, found := document.data[document.IDKey()].(string)
	if found {
		return objID
	}
	return ""
}

func (document *document) Collection() string {
	return document.collection
}
