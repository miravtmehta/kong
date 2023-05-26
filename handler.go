package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/gorilla/mux"
	"net/http"
)

//func getAllServices(w http.ResponseWriter, r *http.Request) {
//	opts := GetOptions(r)
//	data, err := serviceProvider.GetAllService(&opts)
//	if err != nil {
//		fmt.Println(err)
//	}
//	err = json.NewEncoder(w).Encode(data)
//	if err != nil {
//		return
//	}
//}
//
//func getService(w http.ResponseWriter, r *http.Request) {
//	vars := mux.Vars(r)
//	data, err := serviceProvider.GetService(vars["name"])
//	if err != nil {
//		fmt.Println(err)
//	}
//	err = json.NewEncoder(w).Encode(data)
//	if err != nil {
//		return
//	}
//}
//
//func createService(w http.ResponseWriter, r *http.Request) {
//	var service Service
//
//	err := json.NewDecoder(r.Body).Decode(&service)
//	if err != nil {
//		log.Fatalf("There was an error decoding the request body into the struct : %s\n", err.Error())
//	}
//	defer r.Body.Close()
//
//	err = serviceProvider.CreateService(service)
//	if err != nil {
//		log.Fatalf("error creating service: %s\n", err.Error())
//	}
//}

func (a *AppRouter) getAllServices(w http.ResponseWriter, r *http.Request) {
	opts := GetOptions(r)
	data, err := serviceProvider.GetAllService(&opts)
	if err != nil {
		switch err {
		case pg.ErrNoRows:
			msg := fmt.Sprintf("Product not found. Error: %s", err.Error())
			a.respondWithError(w, http.StatusNotFound, msg)
		default:
			a.respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	a.respondWithJSON(w, http.StatusOK, data)
}

func (a *AppRouter) getService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	data, err := serviceProvider.GetService(vars["name"])
	if err != nil {
		switch err {
		case pg.ErrNoRows:
			msg := fmt.Sprintf("Product not found. Error: %s", err.Error())
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

	if err := serviceProvider.CreateService(service); err != nil {
		a.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.respondWithJSON(w, http.StatusCreated, service)

}

func (a *AppRouter) deleteService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := serviceProvider.DeleteService(vars["name"])
	if err != nil {
		switch err {
		case pg.ErrNoRows:
			msg := fmt.Sprintf("Product not found. Error: %s", err.Error())
			a.respondWithError(w, http.StatusBadRequest, msg)
		default:
			a.respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	a.respondWithJSON(w, http.StatusOK, nil)
}

func (a *AppRouter) dumpPG(w http.ResponseWriter, r *http.Request) {
	err := serviceProvider.GenerateRandomPgData()
	if err != nil {
		fmt.Printf("PG DUMP error %s\n", err)
		a.respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	a.respondWithJSON(w, http.StatusCreated, nil)

}
