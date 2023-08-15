package restorer

import "net/http"

// Middleware для синхронизации inmemory хранилища метрик с файлом на диске.
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
