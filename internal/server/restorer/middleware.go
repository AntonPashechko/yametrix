package restorer

import "net/http"

func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Вызов целевого handler
		h.ServeHTTP(w, r)

		//Синхронизируем
		if instance != nil {
			instance.store()
		}
	})
}
