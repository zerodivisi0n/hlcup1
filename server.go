package main

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"github.com/gorilla/schema"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	validator "gopkg.in/go-playground/validator.v9"
	mgo "gopkg.in/mgo.v2"
)

type Store interface {
	// User methods
	CreateUser(u *User) error
	CreateUsers(us []User) error
	UpdateUser(id int, u *User) error
	GetUser(id int, u *User) error
	GetUserVisits(id int, q *UserVisitsQuery, visits *[]UserVisit) error

	// Location methods
	CreateLocation(l *Location) error
	CreateLocations(ls []Location) error
	UpdateLocation(id int, l *Location) error
	GetLocation(id int, l *Location) error
	GetLocationAvg(id int, q *LocationAvgQuery) (float64, error)

	// Visit methods
	CreateVisit(v *Visit) error
	CreateVisits(vs []Visit) error
	UpdateVisit(id int, v *Visit) error
	GetVisit(id int, v *Visit) error

	// Clear the entire databasec
	Clear() error
}

type Server struct {
	store        Store
	validator    *validator.Validate
	queryDecoder *schema.Decoder
}

func NewServer(store Store) *Server {
	v := validator.New()
	v.RegisterCustomTypeFunc(ValidateTimestamp, Timestamp{})
	d := schema.NewDecoder()
	d.IgnoreUnknownKeys(true)
	return &Server{
		store:        store,
		validator:    v,
		queryDecoder: d,
	}
}

func (s *Server) Listen(addr string) error {
	return http.ListenAndServe(addr, s.handler())
}

func (s *Server) handler() http.Handler {
	r := httprouter.New()
	// Users
	// r.POST("/users/new", s.createUser) // conflicts with existing wildcard
	r.POST("/users/:id", s.updateUser)
	r.GET("/users/:id", s.getUser)
	r.GET("/users/:id/visits", s.getUserVisits)

	// Locations
	// r.POST("/locations/new", s.createLocation)
	r.POST("/locations/:id", s.updateLocation)
	r.GET("/locations/:id", s.getLocation)
	r.GET("/locations/:id/avg", s.getLocationAvg)

	// Visits
	// r.POST("/visits/new", s.createVisit)
	r.POST("/visits/:id", s.updateVisit)
	r.GET("/visits/:id", s.getVisit)

	return r
}

// Users endpoints
func (s *Server) createUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.validator.Struct(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.store.CreateUser(&user); err != nil {
		handleDbError(w, err)
		return
	}
	jsonResponse(w, struct{}{})
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if ps.ByName("id") == "new" {
		s.createUser(w, r, ps)
		return
	}
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var user User
	// check user exists first
	if err := s.store.GetUser(id, &user); err != nil {
		handleDbError(w, err)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.validator.Struct(user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.store.UpdateUser(id, &user); err != nil {
		handleDbError(w, err)
		return
	}
	jsonResponse(w, struct{}{})
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var user User
	if err := s.store.GetUser(id, &user); err != nil {
		handleDbError(w, err)
		return
	}
	jsonResponse(w, &user)
}

func (s *Server) getUserVisits(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var query UserVisitsQuery
	if err := s.queryDecoder.Decode(&query, r.URL.Query()); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var visits []UserVisit
	if err := s.store.GetUserVisits(id, &query, &visits); err != nil {
		handleDbError(w, err)
		return
	}
	if len(visits) == 0 {
		visits = make([]UserVisit, 0)
	}
	jsonResponse(w, map[string]interface{}{
		"visits": visits,
	})
}

// Locations endpoints
func (s *Server) createLocation(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var location Location
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.validator.Struct(&location); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.store.CreateLocation(&location); err != nil {
		handleDbError(w, err)
		return
	}
	jsonResponse(w, struct{}{})
}

func (s *Server) updateLocation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if ps.ByName("id") == "new" {
		s.createLocation(w, r, ps)
		return
	}
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var location Location
	// check location exists first
	if err := s.store.GetLocation(id, &location); err != nil {
		handleDbError(w, err)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.validator.Struct(&location); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.store.UpdateLocation(id, &location); err != nil {
		handleDbError(w, err)
		return
	}
	jsonResponse(w, struct{}{})
}

func (s *Server) getLocation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var location Location
	if err := s.store.GetLocation(id, &location); err != nil {
		handleDbError(w, err)
		return
	}
	jsonResponse(w, &location)
}

func (s *Server) getLocationAvg(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var query LocationAvgQuery
	if err := s.queryDecoder.Decode(&query, r.URL.Query()); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.validator.Struct(&query); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	avg, err := s.store.GetLocationAvg(id, &query)
	if err != nil {
		handleDbError(w, err)
		return
	}
	jsonResponse(w, map[string]interface{}{
		"avg": math.Floor(avg*100000+0.5) / 100000,
	})
}

// Visits endpoints
func (s *Server) createVisit(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var visit Visit
	if err := json.NewDecoder(r.Body).Decode(&visit); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.validator.Struct(&visit); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.store.CreateVisit(&visit); err != nil {
		handleDbError(w, err)
		return
	}
	jsonResponse(w, struct{}{})
}

func (s *Server) updateVisit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if ps.ByName("id") == "new" {
		s.createVisit(w, r, ps)
		return
	}
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var visit Visit
	// check location exists first
	if err := s.store.GetVisit(id, &visit); err != nil {
		handleDbError(w, err)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&visit); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.validator.Struct(&visit); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.store.UpdateVisit(id, &visit); err != nil {
		handleDbError(w, err)
		return
	}
	jsonResponse(w, struct{}{})
}

func (s *Server) getVisit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var visit Visit
	if err := s.store.GetVisit(id, &visit); err != nil {
		handleDbError(w, err)
		return
	}
	jsonResponse(w, &visit)
}

func handleDbError(w http.ResponseWriter, err error) {
	if err == mgo.ErrNotFound {
		w.WriteHeader(http.StatusNotFound)
	} else if err == ErrMissingID || err == ErrUpdateID || mgo.IsDup(err) {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		log.Errorf("Database error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func jsonResponse(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(body)
}
