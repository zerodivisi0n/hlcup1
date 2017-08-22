package main

import (
	"encoding/json"
	"math"
	"strconv"

	"github.com/buaazp/fasthttprouter"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
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
	store     Store
	validator *validator.Validate
}

func NewServer(store Store) *Server {
	v := validator.New()
	v.RegisterCustomTypeFunc(ValidateTimestamp, Timestamp{})
	return &Server{
		store:     store,
		validator: v,
	}
}

func (s *Server) Listen(addr string) error {
	return fasthttp.ListenAndServe(addr, s.handler())
}

func (s *Server) handler() fasthttp.RequestHandler {
	r := fasthttprouter.New()
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

	return r.Handler
}

// Users endpoints
func (s *Server) createUser(ctx *fasthttp.RequestCtx) {
	var user User
	ctx.SetConnectionClose()
	if err := json.Unmarshal(ctx.PostBody(), &user); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.validator.Struct(&user); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.store.CreateUser(&user); err != nil {
		handleDbError(ctx, err)
		return
	}
	jsonResponse(ctx, struct{}{})
}

func (s *Server) updateUser(ctx *fasthttp.RequestCtx) {
	if ctx.UserValue("id") == "new" {
		s.createUser(ctx)
		return
	}
	ctx.SetConnectionClose()
	id, err := strconv.Atoi(ctx.UserValue("id").(string))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	var user User
	// check user exists first
	if err := s.store.GetUser(id, &user); err != nil {
		handleDbError(ctx, err)
		return
	}
	if err := json.Unmarshal(ctx.PostBody(), &user); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.validator.Struct(user); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.store.UpdateUser(id, &user); err != nil {
		handleDbError(ctx, err)
		return
	}
	jsonResponse(ctx, struct{}{})
}

func (s *Server) getUser(ctx *fasthttp.RequestCtx) {
	id, err := strconv.Atoi(ctx.UserValue("id").(string))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	var user User
	if err := s.store.GetUser(id, &user); err != nil {
		handleDbError(ctx, err)
		return
	}
	jsonResponse(ctx, &user)
}

func (s *Server) getUserVisits(ctx *fasthttp.RequestCtx) {
	id, err := strconv.Atoi(ctx.UserValue("id").(string))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	var query UserVisitsQuery
	if !parseUserVisitsQuery(ctx.QueryArgs(), &query) {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	var visits []UserVisit
	if err := s.store.GetUserVisits(id, &query, &visits); err != nil {
		handleDbError(ctx, err)
		return
	}
	if len(visits) == 0 {
		visits = make([]UserVisit, 0)
	}
	jsonResponse(ctx, map[string]interface{}{
		"visits": visits,
	})
}

// Locations endpoints
func (s *Server) createLocation(ctx *fasthttp.RequestCtx) {
	var location Location
	ctx.SetConnectionClose()
	if err := json.Unmarshal(ctx.PostBody(), &location); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.validator.Struct(&location); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.store.CreateLocation(&location); err != nil {
		handleDbError(ctx, err)
		return
	}
	jsonResponse(ctx, struct{}{})
}

func (s *Server) updateLocation(ctx *fasthttp.RequestCtx) {
	if ctx.UserValue("id") == "new" {
		s.createLocation(ctx)
		return
	}
	ctx.SetConnectionClose()
	id, err := strconv.Atoi(ctx.UserValue("id").(string))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	var location Location
	// check location exists first
	if err := s.store.GetLocation(id, &location); err != nil {
		handleDbError(ctx, err)
		return
	}
	if err := json.Unmarshal(ctx.PostBody(), &location); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.validator.Struct(&location); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.store.UpdateLocation(id, &location); err != nil {
		handleDbError(ctx, err)
		return
	}
	jsonResponse(ctx, struct{}{})
}

func (s *Server) getLocation(ctx *fasthttp.RequestCtx) {
	id, err := strconv.Atoi(ctx.UserValue("id").(string))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	var location Location
	if err := s.store.GetLocation(id, &location); err != nil {
		handleDbError(ctx, err)
		return
	}
	jsonResponse(ctx, &location)
}

func (s *Server) getLocationAvg(ctx *fasthttp.RequestCtx) {
	id, err := strconv.Atoi(ctx.UserValue("id").(string))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	var query LocationAvgQuery
	if !parseLocationAvgQuery(ctx.QueryArgs(), &query) {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	avg, err := s.store.GetLocationAvg(id, &query)
	if err != nil {
		handleDbError(ctx, err)
		return
	}
	jsonResponse(ctx, map[string]interface{}{
		"avg": math.Floor(avg*100000+0.5) / 100000,
	})
}

// Visits endpoints
func (s *Server) createVisit(ctx *fasthttp.RequestCtx) {
	var visit Visit
	ctx.SetConnectionClose()
	if err := json.Unmarshal(ctx.PostBody(), &visit); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.validator.Struct(&visit); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.store.CreateVisit(&visit); err != nil {
		handleDbError(ctx, err)
		return
	}
	jsonResponse(ctx, struct{}{})
}

func (s *Server) updateVisit(ctx *fasthttp.RequestCtx) {
	if ctx.UserValue("id") == "new" {
		s.createVisit(ctx)
		return
	}
	ctx.SetConnectionClose()
	id, err := strconv.Atoi(ctx.UserValue("id").(string))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	var visit Visit
	// check location exists first
	if err := s.store.GetVisit(id, &visit); err != nil {
		handleDbError(ctx, err)
		return
	}
	if err := json.Unmarshal(ctx.PostBody(), &visit); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.validator.Struct(&visit); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.store.UpdateVisit(id, &visit); err != nil {
		handleDbError(ctx, err)
		return
	}
	jsonResponse(ctx, struct{}{})
}

func (s *Server) getVisit(ctx *fasthttp.RequestCtx) {
	id, err := strconv.Atoi(ctx.UserValue("id").(string))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	var visit Visit
	if err := s.store.GetVisit(id, &visit); err != nil {
		handleDbError(ctx, err)
		return
	}
	jsonResponse(ctx, &visit)
}

func handleDbError(ctx *fasthttp.RequestCtx, err error) {
	if err == mgo.ErrNotFound {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
	} else if err == ErrMissingID || err == ErrUpdateID || mgo.IsDup(err) {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
	} else {
		log.Errorf("Database error: %v", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	}
}

func jsonResponse(ctx *fasthttp.RequestCtx, body interface{}) {
	ctx.SetContentType("application/json; charset=utf-8")
	json.NewEncoder(ctx).Encode(body)
}

func parseUserVisitsQuery(args *fasthttp.Args, q *UserVisitsQuery) bool {
	if val := args.Peek("fromDate"); len(val) > 0 {
		ts, err := strconv.Atoi(string(val))
		if err != nil {
			return false
		}
		q.FromDate.SetUnix(int64(ts))
	}
	if val := args.Peek("toDate"); len(val) > 0 {
		ts, err := strconv.Atoi(string(val))
		if err != nil {
			return false
		}
		q.ToDate.SetUnix(int64(ts))
	}
	q.Country = string(args.Peek("country"))
	if val := args.Peek("toDistance"); len(val) > 0 {
		i, err := strconv.Atoi(string(val))
		if err != nil {
			return false
		}
		q.ToDistance = i
	}

	return true
}

func parseLocationAvgQuery(args *fasthttp.Args, q *LocationAvgQuery) bool {
	if val := args.Peek("fromDate"); len(val) > 0 {
		ts, err := strconv.Atoi(string(val))
		if err != nil {
			return false
		}
		q.FromDate.SetUnix(int64(ts))
	}
	if val := args.Peek("toDate"); len(val) > 0 {
		ts, err := strconv.Atoi(string(val))
		if err != nil {
			return false
		}
		q.ToDate.SetUnix(int64(ts))
	}
	if val := args.Peek("fromAge"); len(val) > 0 {
		i, err := strconv.Atoi(string(val))
		if err != nil {
			return false
		}
		q.FromAge = i
	}
	if val := args.Peek("toAge"); len(val) > 0 {
		i, err := strconv.Atoi(string(val))
		if err != nil {
			return false
		}
		q.ToAge = i
	}
	q.Gender = string(args.Peek("gender"))
	if q.Gender != "" && q.Gender != "m" && q.Gender != "f" {
		return false
	}
	return true
}
