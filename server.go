package main

import (
	"bytes"
	"errors"
	"math"

	"github.com/buger/jsonparser"
	"github.com/mailru/easyjson"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

var emptyResponseBody = []byte("{}\n")

var (
	ErrMissingID = errors.New("missing id")
	ErrNotFound  = errors.New("not found")
	ErrUpdateID  = errors.New("id field cannot be changed")
	ErrDup       = errors.New("duplicate key error")
)

type Store interface {
	// User methods
	CreateUser(u *User) error
	CreateUsers(us []User) error
	UpdateUser(id uint, u *User) error
	GetUser(id uint, u *User) error
	GetUserVisits(id uint, q *UserVisitsQuery, visits *[]UserVisit) error

	// Location methods
	CreateLocation(l *Location) error
	CreateLocations(ls []Location) error
	UpdateLocation(id uint, l *Location) error
	GetLocation(id uint, l *Location) error
	GetLocationAvg(id uint, q *LocationAvgQuery) (float64, error)

	// Visit methods
	CreateVisit(v *Visit) error
	CreateVisits(vs []Visit) error
	UpdateVisit(id uint, v *Visit) error
	GetVisit(id uint, v *Visit) error

	// Clear the entire databasec
	Clear() error
}

type Server struct {
	store Store
}

func NewServer(store Store) *Server {
	return &Server{
		store: store,
	}
}

func (s *Server) Listen(addr string) error {
	return fasthttp.ListenAndServe(addr, s.handler)
}

func (s *Server) handler(ctx *fasthttp.RequestCtx) {
	path := ctx.Path()
	if ctx.IsPost() {
		if bytes.Equal(path, []byte("/users/new")) {
			s.createUser(ctx)
		} else if bytes.HasPrefix(path, []byte("/users/")) {
			s.updateUser(ctx)
		} else if bytes.Equal(path, []byte("/locations/new")) {
			s.createLocation(ctx)
		} else if bytes.HasPrefix(path, []byte("/locations/")) {
			s.updateLocation(ctx)
		} else if bytes.Equal(path, []byte("/visits/new")) {
			s.createVisit(ctx)
		} else if bytes.HasPrefix(path, []byte("/visits/")) {
			s.updateVisit(ctx)
		} else {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}
	} else if ctx.IsGet() {
		if bytes.HasPrefix(path, []byte("/users/")) {
			if bytes.HasSuffix(path, []byte("/visits")) {
				s.getUserVisits(ctx)
			} else {
				s.getUser(ctx)
			}
		} else if bytes.HasPrefix(path, []byte("/locations/")) {
			if bytes.HasSuffix(path, []byte("/avg")) {
				s.getLocationAvg(ctx)
			} else {
				s.getLocation(ctx)
			}
		} else if bytes.HasPrefix(path, []byte("/visits/")) {
			s.getVisit(ctx)
		} else {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}
	} else {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
	}
}

// Users endpoints
func (s *Server) createUser(ctx *fasthttp.RequestCtx) {
	var user User
	ctx.SetConnectionClose()
	if err := user.UnmarshalJSON(ctx.PostBody()); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if !user.Validate() {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.store.CreateUser(&user); err != nil {
		handleDbError(ctx, err)
		return
	}
	emptyResponse(ctx)
}

func (s *Server) updateUser(ctx *fasthttp.RequestCtx) {
	ctx.SetConnectionClose()
	id, err := jsonparser.ParseInt(ctx.Path()[7:])
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	var user User
	// check user exists first
	if err := s.store.GetUser(uint(id), &user); err != nil {
		handleDbError(ctx, err)
		return
	}
	if err := user.UnmarshalJSON(ctx.PostBody()); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if !user.Validate() {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.store.UpdateUser(uint(id), &user); err != nil {
		handleDbError(ctx, err)
		return
	}
	emptyResponse(ctx)
}

func (s *Server) getUser(ctx *fasthttp.RequestCtx) {
	id, err := jsonparser.ParseInt(ctx.Path()[7:])
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	var user User
	if err := s.store.GetUser(uint(id), &user); err != nil {
		handleDbError(ctx, err)
		return
	}
	jsonResponse(ctx, &user)
}

func (s *Server) getUserVisits(ctx *fasthttp.RequestCtx) {
	id, err := jsonparser.ParseInt(ctx.Path()[7 : len(ctx.Path())-7])
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
	if err := s.store.GetUserVisits(uint(id), &query, &visits); err != nil {
		handleDbError(ctx, err)
		return
	}
	if len(visits) == 0 {
		visits = make([]UserVisit, 0)
	}
	jsonResponse(ctx, &UserVisitsResult{visits})
}

// Locations endpoints
func (s *Server) createLocation(ctx *fasthttp.RequestCtx) {
	var location Location
	ctx.SetConnectionClose()
	if err := location.UnmarshalJSON(ctx.PostBody()); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if !location.Validate() {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.store.CreateLocation(&location); err != nil {
		handleDbError(ctx, err)
		return
	}
	emptyResponse(ctx)
}

func (s *Server) updateLocation(ctx *fasthttp.RequestCtx) {
	ctx.SetConnectionClose()
	id, err := jsonparser.ParseInt(ctx.Path()[11:])
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	var location Location
	// check location exists first
	if err := s.store.GetLocation(uint(id), &location); err != nil {
		handleDbError(ctx, err)
		return
	}
	if err := location.UnmarshalJSON(ctx.PostBody()); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if !location.Validate() {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.store.UpdateLocation(uint(id), &location); err != nil {
		handleDbError(ctx, err)
		return
	}
	emptyResponse(ctx)
}

func (s *Server) getLocation(ctx *fasthttp.RequestCtx) {
	id, err := jsonparser.ParseInt(ctx.Path()[11:])
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	var location Location
	if err := s.store.GetLocation(uint(id), &location); err != nil {
		handleDbError(ctx, err)
		return
	}
	jsonResponse(ctx, &location)
}

func (s *Server) getLocationAvg(ctx *fasthttp.RequestCtx) {
	id, err := jsonparser.ParseInt(ctx.Path()[11 : len(ctx.Path())-4])
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	var query LocationAvgQuery
	if !parseLocationAvgQuery(ctx.QueryArgs(), &query) {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	avg, err := s.store.GetLocationAvg(uint(id), &query)
	if err != nil {
		handleDbError(ctx, err)
		return
	}
	result := LocationAvgResult{
		Avg: math.Floor(avg*100000+0.5) / 100000,
	}
	jsonResponse(ctx, &result)
}

// Visits endpoints
func (s *Server) createVisit(ctx *fasthttp.RequestCtx) {
	var visit Visit
	ctx.SetConnectionClose()
	if err := visit.UnmarshalJSON(ctx.PostBody()); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if !visit.Validate() {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.store.CreateVisit(&visit); err != nil {
		handleDbError(ctx, err)
		return
	}
	emptyResponse(ctx)
}

func (s *Server) updateVisit(ctx *fasthttp.RequestCtx) {
	ctx.SetConnectionClose()
	id, err := jsonparser.ParseInt(ctx.Path()[8:])
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	var visit Visit
	// check location exists first
	if err := s.store.GetVisit(uint(id), &visit); err != nil {
		handleDbError(ctx, err)
		return
	}
	if err := visit.UnmarshalJSON(ctx.PostBody()); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if !visit.Validate() {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if err := s.store.UpdateVisit(uint(id), &visit); err != nil {
		handleDbError(ctx, err)
		return
	}
	emptyResponse(ctx)
}

func (s *Server) getVisit(ctx *fasthttp.RequestCtx) {
	id, err := jsonparser.ParseInt(ctx.Path()[8:])
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	var visit Visit
	if err := s.store.GetVisit(uint(id), &visit); err != nil {
		handleDbError(ctx, err)
		return
	}
	jsonResponse(ctx, &visit)
}

func handleDbError(ctx *fasthttp.RequestCtx, err error) {
	if err == ErrNotFound {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
	} else if err == ErrMissingID || err == ErrUpdateID || err == ErrDup {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
	} else {
		log.Errorf("Database error: %v", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	}
}

func jsonResponse(ctx *fasthttp.RequestCtx, body easyjson.Marshaler) {
	ctx.SetContentType("application/json; charset=utf-8")
	easyjson.MarshalToWriter(body, ctx)
}

func emptyResponse(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json; charset=utf-8")
	ctx.Write(emptyResponseBody)
}

func parseUserVisitsQuery(args *fasthttp.Args, q *UserVisitsQuery) bool {
	if val := args.Peek("fromDate"); len(val) > 0 {
		ts, err := jsonparser.ParseInt(val)
		if err != nil {
			return false
		}
		q.FromDate = &ts
	}
	if val := args.Peek("toDate"); len(val) > 0 {
		ts, err := jsonparser.ParseInt(val)
		if err != nil {
			return false
		}
		q.ToDate = &ts
	}
	q.Country = string(args.Peek("country"))
	if val := args.Peek("toDistance"); len(val) > 0 {
		i, err := jsonparser.ParseInt(val)
		if err != nil {
			return false
		}
		ii := int(i)
		q.ToDistance = &ii
	}

	return true
}

func parseLocationAvgQuery(args *fasthttp.Args, q *LocationAvgQuery) bool {
	if val := args.Peek("fromDate"); len(val) > 0 {
		ts, err := jsonparser.ParseInt(val)
		if err != nil {
			return false
		}
		q.FromDate = &ts
	}
	if val := args.Peek("toDate"); len(val) > 0 {
		ts, err := jsonparser.ParseInt(val)
		if err != nil {
			return false
		}
		q.ToDate = &ts
	}
	if val := args.Peek("fromAge"); len(val) > 0 {
		i, err := jsonparser.ParseInt(val)
		if err != nil {
			return false
		}
		ii := int(i)
		q.FromAge = &ii
	}
	if val := args.Peek("toAge"); len(val) > 0 {
		i, err := jsonparser.ParseInt(val)
		if err != nil {
			return false
		}
		ii := int(i)
		q.ToAge = &ii
	}
	q.Gender = string(args.Peek("gender"))
	if q.Gender != "" && q.Gender != "m" && q.Gender != "f" {
		return false
	}
	return true
}
