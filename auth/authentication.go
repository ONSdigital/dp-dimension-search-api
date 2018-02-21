package auth

import (
	"errors"
	"net/http"

	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/dp-search-api/models"
)

// Authenticator structure which holds the secret key for validating clients. This will be replaced in the future, after the `thin-slices` has been delivered
type Authenticator struct {
	SecretKey  string
	HeaderName string
	Subnet     string
}

// Check wraps a HTTP handle. If authentication fails an error code is returned else the HTTP handler is called
func (a *Authenticator) Check(handle func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// If we're not in in the publishing subnet. Nothing has auth
		if a.Subnet != models.SubnetPublishing {
			http.Error(w, "Unauthorised access to API", http.StatusUnauthorized)
			log.Error(errors.New("attempted access to authorisation required endpoint via the web-subnet"), log.Data{"header": a.HeaderName})
			return
		} else {

			// Otherwise it depends on their credentials
			key := r.Header.Get(a.HeaderName)
			if key == "" {
				http.Error(w, "No authentication header provided", http.StatusUnauthorized)
				log.Error(errors.New("client missing token"), log.Data{"header": a.HeaderName})
				return
			}
			if key != a.SecretKey {
				http.Error(w, "Unauthorised access to API", http.StatusUnauthorized)
				log.Error(errors.New("unauthorised access to API"), log.Data{"header": a.HeaderName})
				return
			}

			// The request has been authenticated, now run the clients request
			handle(w, r)
		}
	})
}

// ManualCheck sets `isAuthenticated` and passes that to the HTTP handler, the handler is responsible for setting the status code
func (a *Authenticator) ManualCheck(handle func(http.ResponseWriter, *http.Request, bool)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// If we're not in in the publishing subnet. Nothing has auth
		if a.Subnet != models.SubnetPublishing {
			handle(w, r, false)
		} else {

			// Otherwise it depends on their credentials
			isAuthenticated := false
			key := r.Header.Get(a.HeaderName)
			if key == a.SecretKey {
				isAuthenticated = true
			}
			handle(w, r, isAuthenticated)
		}
	})
}
