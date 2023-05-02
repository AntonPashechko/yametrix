package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	once     sync.Once
	compress bool

	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

/*Проверим, что контент application/json или text/html - будем пресовать*/
func (c *compressWriter) needCompress() bool {
	c.once.Do(func() {
		for _, v := range c.w.Header()["Content-Type"] {
			if v == "application/json" || v == "text/html" {
				c.w.Header().Set("Content-Encoding", "gzip")
				c.compress = true
			}
		}
	})

	return c.compress
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if c.needCompress() {
		return c.zw.Write(p)
	} else {
		return c.w.Write(p)
	}
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	if c.needCompress() {
		return c.zw.Close()
	} else {
		return nil
	}
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ow := w
		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := newCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	})
}
