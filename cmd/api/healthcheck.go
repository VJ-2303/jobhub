package main

import "net/http"

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"status":      "OK",
		"environment": app.config.environment,
		"version":     "1.0.0",
	}
	app.writeJSON(w, http.StatusOK, envelope{"health": data}, nil)
}
