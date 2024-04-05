package jsonx

import (
	"errors"
	"io"
	"net/http"

	"code-platform/pkg/strconvx"

	"github.com/bytedance/sonic/decoder"
	"github.com/gin-gonic/gin/binding"
)

type jsonxBinding struct{}

var SonicDecoder binding.Binding = jsonxBinding{}

func (s jsonxBinding) Name() string {
	return "jsonx"
}

func (s jsonxBinding) Bind(req *http.Request, obj interface{}) error {
	if req == nil || req.Body == nil {
		return errors.New("invalid request")
	}
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	dec := decoder.NewDecoder(strconvx.BytesToString(data))
	return dec.Decode(obj)
}
