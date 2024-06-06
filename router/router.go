package routes

import (
	"banking/controllers"
	"github.com/gorilla/mux"
	"net/http"
)


// Middleware function to check if the password is correct
func passwordMiddleware(password string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pass := r.URL.Query().Get("password")
			if pass != password {
				http.Error(w, "Incorrect password", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Routers(password string) *mux.Router {
	router := mux.NewRouter()

	router.Use(passwordMiddleware(password))
	
	router.HandleFunc("/account", controllers.CreateAccountHandler).Methods("POST")
	router.HandleFunc("/accounts", controllers.GetAllAccountHandler).Methods("GET")
	router.HandleFunc("/account/{id}", controllers.GetOneAccountHandler).Methods("GET")
	router.HandleFunc("/account/{id}", controllers.DeleteAccountHandler).Methods("DELETE")
	router.HandleFunc("/accounts", controllers.DeleteAllAccountsHandler).Methods("DELETE")
	router.HandleFunc("/payments", controllers.PaymentHandler).Methods("POST")
	return router
}
