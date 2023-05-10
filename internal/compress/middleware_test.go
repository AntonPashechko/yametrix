package compress

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, body []byte, gzipReq bool) *resty.Response {

	req := resty.New().R()
	req.Method = "POST"
	req.URL = ts.URL + "/test"
	req.SetHeader("Content-Type", "text/html")

	content := body
	if gzipReq {
		content, _ = GzipCompress(body)
		req.SetHeader("Content-Encoding", "gzip")
	}
	req.SetBody(content)

	resp, err := req.Send()
	require.NoError(t, err)
	return resp
}

func TestMiddleware(t *testing.T) {
	//Этот хэндлер будет за мидлеваре идти и добавит в полученную строку Hello, %s
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		b, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		resp := fmt.Sprintf("Hello, %s!", string(b))
		w.Write([]byte(resp))
	})

	handlerToTest := Middleware(nextHandler)

	ts := httptest.NewServer(handlerToTest)
	defer ts.Close()

	tests := []struct {
		name string
		gzip bool
	}{
		{"test no gzip", false},
		{"test gzip", true},

		//На вервер всегда прилетает заголовок ("Accept-Encoding", "gzip"), что бы я не делал, как это отключить?

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := testRequest(t, ts, []byte("Mike"), tt.gzip)
			assert.Equal(t, http.StatusOK, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")

			assert.Equal(t, "Hello, Mike!", string(resp.Body()), "Содержимое ответа не совпадает с ожидаемым")
		})
	}
}
