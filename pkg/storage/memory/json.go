package memory

import (
	"encoding/json"
	"os"
)

type Persistence interface {
	Save(v interface{}) error
	Load(v interface{}) error
}

func JSON(fn string) Persistence {
	return &jsonPersistence{fn}
}

type jsonPersistence struct {
	fn string
}

func (j *jsonPersistence) Save(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return os.WriteFile(j.fn, data, 0666)
}

func (j *jsonPersistence) Load(v interface{}) error {
	data, err := os.ReadFile(j.fn)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}
