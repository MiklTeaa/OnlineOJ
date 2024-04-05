package jsonx

import (
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin/render"
)

type SonicEncoder struct {
	Data interface{}
}

func NewSonicEncoder(data interface{}) SonicEncoder {
	return SonicEncoder{Data: data}
}

var _ render.Render = SonicEncoder{}

func (s SonicEncoder) Render(w http.ResponseWriter) error {
	s.WriteContentType(w)
	data, err := sonic.Marshal(s.Data)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

var jsonContentType = []string{"application/json; charset=utf-8"}

func (s SonicEncoder) WriteContentType(w http.ResponseWriter) {
	w.Header()["Content-Type"] = jsonContentType
}
