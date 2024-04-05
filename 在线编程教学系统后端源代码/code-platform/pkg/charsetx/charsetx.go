package charsetx

import (
	"io"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func GBKToUTF8(reader io.Reader) io.Reader {
	return transform.NewReader(reader, simplifiedchinese.GBK.NewDecoder())
}
