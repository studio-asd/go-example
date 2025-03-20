// We have some different behavior for web and mobile application, where in web application we can simply redirect the
// user to the page that we want to show, but in mobile application we need to handle the navigation in the application.

package services

import "net/http"

func handleWebAuthentication(w http.ResponseWriter, r *http.Request) {
}
