package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/gorilla/mux"
	"net/http"
)

func (a *AppRouter) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *AppRouter) respondWithError(w http.ResponseWriter, code int, message string) {
	a.respondWithJSON(w, code, map[string]string{"error": message})
	a.logger.Printf("App error: code %d, message %s", code, message)
}

func (a *AppRouter) getAllServices(w http.ResponseWriter, r *http.Request) {
	opts := GetOptions(r)
	data, err := a.serviceProvider.GetAllService(&opts)
	if err != nil {
		switch err {
		case pg.ErrNoRows:
			msg := fmt.Sprintf("services not found. Error: %s", err.Error())
			a.respondWithError(w, http.StatusBadRequest, msg)
		default:
			a.respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	a.respondWithJSON(w, http.StatusOK, data)
}

func (a *AppRouter) getService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	data, err := a.serviceProvider.GetService(vars["name"])
	if err != nil {
		switch err {
		case pg.ErrNoRows:
			msg := fmt.Sprintf("service %s not found. Error: %s", vars["name"], err.Error())
			a.respondWithError(w, http.StatusBadRequest, msg)
		default:
			a.respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	a.respondWithJSON(w, http.StatusOK, data)
}

func (a *AppRouter) createService(w http.ResponseWriter, r *http.Request) {
	var service Service
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&service); err != nil {
		msg := fmt.Sprintf("Invalid request payload. Error: %s", err.Error())
		a.respondWithError(w, http.StatusBadRequest, msg)
		return
	}
	defer r.Body.Close()

	if err := a.serviceProvider.CreateService(service); err != nil {
		a.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.respondWithJSON(w, http.StatusCreated, service)

}

func (a *AppRouter) deleteService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := a.serviceProvider.DeleteService(vars["name"])
	if err != nil {
		switch err {
		case pg.ErrNoRows:
			msg := fmt.Sprintf("service %s not found. Error: %s", vars["name"], err.Error())
			a.respondWithError(w, http.StatusBadRequest, msg)
		default:
			a.respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	a.respondWithJSON(w, http.StatusOK, nil)
}

func (a *AppRouter) dump(w http.ResponseWriter, r *http.Request) {
	err := a.serviceProvider.GenerateRandomPgData()
	if err != nil {
		a.respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	a.respondWithJSON(w, http.StatusCreated, nil)
}

func (a *AppRouter) cleanDump(w http.ResponseWriter, r *http.Request) {
	err := a.serviceProvider.DeleteService("*")

	if err != nil {
		a.respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	a.respondWithJSON(w, http.StatusCreated, nil)
}
