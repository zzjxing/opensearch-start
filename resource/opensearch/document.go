package opensearch

import (
	"encoding/json"
	"fmt"
	"time"
)

type Document interface {
	GetID() string
	Bytes() ([]byte, error)
}
type VectorDoc struct {
	ID          string    `json:"id"`
	CreatedTime time.Time `json:"created_time"`
	UpdatedTime time.Time `json:"updated_time"`
	Embedding   []float64 `json:"embedding"`
}

func NewDocument(id string, embedding []float64) Document {
	return VectorDoc{
		ID:          id,
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
		Embedding:   embedding,
	}
}

func (c VectorDoc) GetID() string {
	return c.ID
}

func (c VectorDoc) Bytes() ([]byte, error) {
	body, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func GetVecDocConfig(numberOfReplicas, numberOfShards, dimension int) string {
	baseIndex := `{
		"settings": {
			"number_of_replicas": %d,
			"number_of_shards": %d,
			"index.knn": true
		},
		"mappings": {
			"properties": {
				"id": {
					"type": "keyword"
				},
				"created_time": {
					"type": "date"
				},
				"updated_time": {
					"type": "date"
				},
				"embedding": {
					"type": "knn_vector",
					"dimension": %d
				}
			}
		}
	}`
	return fmt.Sprintf(baseIndex, numberOfReplicas, numberOfShards, dimension)
}
