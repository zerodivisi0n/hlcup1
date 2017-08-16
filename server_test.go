package main

import (
	"errors"
	"io/ioutil"
	"net"
	"testing"
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

func TestHandlers(t *testing.T) {
	type StoreMethod struct {
		method     string
		args       []interface{}
		returnArgs []interface{}
		run        func(args mock.Arguments)
	}
	tt := []struct {
		name         string
		path         string
		request      string
		query        string
		response     string
		statusCode   int
		storeMethods []StoreMethod
	}{
		//--------------------------------
		// User endpoints tests
		//--------------------------------
		{
			name:     "CreateUser",
			path:     "/users/new",
			request:  `{"id":1,"first_name":"First","last_name":"User","email":"foo@bar.com","gender":"m","birth_date":100000}`,
			response: "{}\n",
			storeMethods: []StoreMethod{
				{
					method:     "CreateUser",
					args:       []interface{}{mock.AnythingOfType("*main.User")},
					returnArgs: []interface{}{nil},
				},
			},
		},
		{
			name:       "CreateUser/InvalidBody",
			path:       "/users/new",
			request:    `{bad-json}`,
			statusCode: fasthttp.StatusBadRequest,
		},
		{
			name:       "CreateUser/WithoutID",
			path:       "/users/new",
			request:    `{"first_name":"User"}`,
			statusCode: fasthttp.StatusBadRequest,
			storeMethods: []StoreMethod{
				{
					method:     "CreateUser",
					args:       []interface{}{mock.AnythingOfType("*main.User")},
					returnArgs: []interface{}{ErrMissingID},
				},
			},
		},
		{
			name:       "CreateUser/ValidationError",
			path:       "/users/new",
			request:    `{"first_name":"Alien","gender":"u"}`,
			statusCode: fasthttp.StatusBadRequest,
		},
		{
			name:       "CreateUser/DatabaseError",
			path:       "/users/new",
			request:    `{"id":1,"first_name":"First","last_name":"User","email":"foo@bar.com","gender":"m","birth_date":100000}`,
			statusCode: fasthttp.StatusInternalServerError,
			storeMethods: []StoreMethod{
				{
					method:     "CreateUser",
					args:       []interface{}{mock.AnythingOfType("*main.User")},
					returnArgs: []interface{}{errors.New("db error")},
				},
			},
		},
		{
			name:       "CreateUser/NonUniqEmail",
			path:       "/users/new",
			request:    `{"id":1,"first_name":"First","last_name":"User","email":"duplicate@email.com","gender":"m","birth_date":100000}`,
			statusCode: fasthttp.StatusBadRequest,
			storeMethods: []StoreMethod{
				{
					method:     "CreateUser",
					args:       []interface{}{mock.AnythingOfType("*main.User")},
					returnArgs: []interface{}{&mgo.LastError{Code: 11000}},
				},
			},
		},
		{
			name:       "CreateUser/WithNullField",
			path:       "/users/new",
			request:    `{"id":1,"first_name":"First","last_name":"User","email":null,"gender":"m","birth_date":100000}`,
			statusCode: fasthttp.StatusBadRequest,
		},
		{
			name:     "UpdateUser",
			path:     "/users/1",
			request:  `{"first_name":"Updated"}`,
			response: "{}\n",
			storeMethods: []StoreMethod{
				{
					method:     "GetUser",
					args:       []interface{}{1, mock.AnythingOfType("*main.User")},
					returnArgs: []interface{}{nil},
					run: func(args mock.Arguments) {
						u := args.Get(1).(*User)
						u.LastName = "LastName"
					},
				},
				{
					method:     "UpdateUser",
					args:       []interface{}{1, mock.AnythingOfType("*main.User")},
					returnArgs: []interface{}{nil},
					run: func(args mock.Arguments) {
						u := args.Get(1).(*User)
						assert.Equal(t, "Updated", u.FirstName)
						assert.Equal(t, "LastName", u.LastName)
					},
				},
			},
		},
		{
			name:       "UpdateUser/InvalidID",
			path:       "/users/a",
			statusCode: fasthttp.StatusNotFound,
		},
		{
			name:       "UpdateUser/NotFound",
			path:       "/users/2",
			statusCode: fasthttp.StatusNotFound,
			storeMethods: []StoreMethod{
				{
					method:     "GetUser",
					args:       []interface{}{2, mock.AnythingOfType("*main.User")},
					returnArgs: []interface{}{mgo.ErrNotFound},
				},
			},
		},
		{
			name:       "UpdateUser/InvalidBody",
			path:       "/users/1",
			statusCode: fasthttp.StatusBadRequest,
			request:    `{bad-json}`,
			storeMethods: []StoreMethod{
				{
					method:     "GetUser",
					args:       []interface{}{1, mock.AnythingOfType("*main.User")},
					returnArgs: []interface{}{nil},
				},
			},
		},
		{
			name:       "UpdateUser/ValidationError",
			path:       "/users/1",
			statusCode: fasthttp.StatusBadRequest,
			request:    `{"first_name":"Alien","gender":"u"}`,
			storeMethods: []StoreMethod{
				{
					method:     "GetUser",
					args:       []interface{}{1, mock.AnythingOfType("*main.User")},
					returnArgs: []interface{}{nil},
				},
			},
		},
		{
			name:       "UpdateUser/ErrUpdateID",
			path:       "/users/1",
			statusCode: fasthttp.StatusBadRequest,
			request:    `{"id":2,"email": "new@email.com"}`,
			storeMethods: []StoreMethod{
				{
					method:     "GetUser",
					args:       []interface{}{1, mock.AnythingOfType("*main.User")},
					returnArgs: []interface{}{nil},
				},
				{
					method:     "UpdateUser",
					args:       []interface{}{1, mock.AnythingOfType("*main.User")},
					returnArgs: []interface{}{ErrUpdateID},
				},
			},
		},
		{
			name: "UpdateUser/WithNullField",
			path: "/users/1",
			request: `{
				"email": null
			}`,
			statusCode: fasthttp.StatusBadRequest,
			storeMethods: []StoreMethod{
				{
					method:     "GetUser",
					args:       []interface{}{1, mock.AnythingOfType("*main.User")},
					returnArgs: []interface{}{nil},
					run: func(args mock.Arguments) {
						user := args.Get(1).(*User)
						*user = User{
							ID:        1,
							FirstName: "First",
							LastName:  "User",
							Email:     "foo@bar.com",
							Gender:    "m",
							BirthDate: Timestamp{time.Unix(100000, 0)},
						}
					},
				},
			},
		},
		{
			name:     "GetUser",
			path:     "/users/1",
			response: `{"id":1,"first_name":"First","last_name":"User","email":"foo@bar.com","gender":"m","birth_date":100000}` + "\n",
			storeMethods: []StoreMethod{
				{
					method:     "GetUser",
					args:       []interface{}{1, mock.AnythingOfType("*main.User")},
					returnArgs: []interface{}{nil},
					run: func(args mock.Arguments) {
						user := args.Get(1).(*User)
						*user = User{
							ID:        1,
							FirstName: "First",
							LastName:  "User",
							Email:     "foo@bar.com",
							Gender:    "m",
							BirthDate: Timestamp{time.Unix(100000, 0)},
						}
					},
				},
			},
		},
		{
			name:       "GetUser/InvalidID",
			path:       "/users/a",
			statusCode: fasthttp.StatusNotFound,
		},
		{
			name:       "GetUser/NotFound",
			path:       "/users/1",
			statusCode: fasthttp.StatusNotFound,
			storeMethods: []StoreMethod{
				{
					method:     "GetUser",
					args:       []interface{}{1, mock.AnythingOfType("*main.User")},
					returnArgs: []interface{}{mgo.ErrNotFound},
				},
			},
		},
		{
			name:     "GetUserVisits",
			path:     "/users/1/visits",
			response: `{"visits":[{"mark":5,"visited_at":5000000,"place":"First Place"},{"mark":3,"visited_at":20732957,"place":"Another Place"}]}` + "\n",
			storeMethods: []StoreMethod{
				{
					method:     "GetUserVisits",
					args:       []interface{}{1, mock.AnythingOfType("*main.UserVisitsQuery"), mock.AnythingOfType("*[]main.UserVisit")},
					returnArgs: []interface{}{nil},
					run: func(args mock.Arguments) {
						visits := args.Get(2).(*[]UserVisit)
						*visits = []UserVisit{
							{
								Mark:      5,
								VisitedAt: Timestamp{time.Unix(5000000, 0)},
								Place:     "First Place",
							},
							{
								Mark:      3,
								VisitedAt: Timestamp{time.Unix(20732957, 0)},
								Place:     "Another Place",
							},
						}
					},
				},
			},
		},
		{
			name:       "GetUserVisits/InvalidID",
			path:       "/users/a/visits",
			statusCode: fasthttp.StatusNotFound,
		},
		{
			name:       "GetUserVisits/NotFound",
			path:       "/users/999/visits",
			statusCode: fasthttp.StatusNotFound,
			storeMethods: []StoreMethod{
				{
					method:     "GetUserVisits",
					args:       []interface{}{999, mock.AnythingOfType("*main.UserVisitsQuery"), mock.AnythingOfType("*[]main.UserVisit")},
					returnArgs: []interface{}{mgo.ErrNotFound},
				},
			},
		},
		{
			name:     "GetUserVisits/EmptyResults",
			path:     "/users/2/visits",
			response: `{"visits":[]}` + "\n",
			storeMethods: []StoreMethod{
				{
					method:     "GetUserVisits",
					args:       []interface{}{2, mock.AnythingOfType("*main.UserVisitsQuery"), mock.AnythingOfType("*[]main.UserVisit")},
					returnArgs: []interface{}{nil},
				},
			},
		},
		{
			name:     "GetUserVisits/WithQuery",
			path:     "/users/1/visits",
			query:    "?fromDate=53636439",
			response: `{"visits":[]}` + "\n",
			storeMethods: []StoreMethod{
				{
					method:     "GetUserVisits",
					args:       []interface{}{1, mock.AnythingOfType("*main.UserVisitsQuery"), mock.AnythingOfType("*[]main.UserVisit")},
					returnArgs: []interface{}{nil},
					run: func(args mock.Arguments) {
						query := args.Get(1).(*UserVisitsQuery)
						assert.Equal(t, Timestamp{time.Unix(53636439, 0)}, query.FromDate)
					},
				},
			},
		},
		{
			name:       "GetUserVisits/WithInvalidQuery",
			path:       "/users/1/visits",
			query:      "?toDate=a",
			statusCode: fasthttp.StatusBadRequest,
		},
		{
			name:     "GetUserVisits/WithUnknownQuery",
			path:     "/users/1/visits",
			query:    "?unknown=value",
			response: `{"visits":[]}` + "\n",
			storeMethods: []StoreMethod{
				{
					method:     "GetUserVisits",
					args:       []interface{}{1, &UserVisitsQuery{}, mock.AnythingOfType("*[]main.UserVisit")},
					returnArgs: []interface{}{nil},
				},
			},
		},
		//------------------------------
		// Location endpoints tests
		//------------------------------
		{
			name:     "CreateLocation",
			path:     "/locations/new",
			request:  `{"id":1,"city":"Moscow","country":"Russia","place":"Red Square","distance":25}`,
			response: "{}\n",
			storeMethods: []StoreMethod{
				{
					method:     "CreateLocation",
					args:       []interface{}{mock.AnythingOfType("*main.Location")},
					returnArgs: []interface{}{nil},
				},
			},
		},
		{
			name:       "CreateLocation/InvalidBody",
			path:       "/locations/new",
			request:    `{bad-json}`,
			statusCode: fasthttp.StatusBadRequest,
		},
		{
			name:       "CreateLocation/ValidationError",
			path:       "/locations/new",
			request:    `{"distance":-5}`,
			statusCode: fasthttp.StatusBadRequest,
		},
		{
			name:       "CreateLocation/WithNullField",
			path:       "/locations/new",
			request:    `{"id":1,"city":"Moscow","country":"Russia","place":"Some Place","distance":null}`,
			statusCode: fasthttp.StatusBadRequest,
		},
		{
			name:       "CreateUser/DatabaseError",
			path:       "/locations/new",
			request:    `{"id":1,"city":"Moscow","country":"Russia","place":"Some Place","distance":25}`,
			statusCode: fasthttp.StatusInternalServerError,
			storeMethods: []StoreMethod{
				{
					method:     "CreateLocation",
					args:       []interface{}{mock.AnythingOfType("*main.Location")},
					returnArgs: []interface{}{errors.New("db error")},
				},
			},
		},
		{
			name:     "UpdateLocation",
			path:     "/locations/1",
			request:  `{"place":"Another place"}`,
			response: "{}\n",
			storeMethods: []StoreMethod{
				{
					method:     "GetLocation",
					args:       []interface{}{1, mock.AnythingOfType("*main.Location")},
					returnArgs: []interface{}{nil},
					run: func(args mock.Arguments) {
						l := args.Get(1).(*Location)
						l.Country = "Russia"
					},
				},
				{
					method:     "UpdateLocation",
					args:       []interface{}{1, mock.AnythingOfType("*main.Location")},
					returnArgs: []interface{}{nil},
					run: func(args mock.Arguments) {
						l := args.Get(1).(*Location)
						assert.Equal(t, "Another place", l.Place)
						assert.Equal(t, "Russia", l.Country)
					},
				},
			},
		},
		{
			name:       "UpdateLocation/InvalidID",
			path:       "/locations/a",
			statusCode: fasthttp.StatusNotFound,
		},
		{
			name:       "UpdateLocation/NotFound",
			path:       "/locations/2",
			statusCode: fasthttp.StatusNotFound,
			storeMethods: []StoreMethod{
				{
					method:     "GetLocation",
					args:       []interface{}{2, mock.AnythingOfType("*main.Location")},
					returnArgs: []interface{}{mgo.ErrNotFound},
				},
			},
		},
		{
			name:       "UpdateLocation/InvalidBody",
			path:       "/locations/1",
			statusCode: fasthttp.StatusBadRequest,
			request:    `{bad-json}`,
			storeMethods: []StoreMethod{
				{
					method:     "GetLocation",
					args:       []interface{}{1, mock.AnythingOfType("*main.Location")},
					returnArgs: []interface{}{nil},
				},
			},
		},
		{
			name:       "UpdateLocation/ValidationError",
			path:       "/locations/1",
			statusCode: fasthttp.StatusBadRequest,
			request:    `{"distance":-50}`,
			storeMethods: []StoreMethod{
				{
					method:     "GetLocation",
					args:       []interface{}{1, mock.AnythingOfType("*main.Location")},
					returnArgs: []interface{}{nil},
				},
			},
		},
		{
			name:       "UpdateLocation/WithNullField",
			path:       "/locations/1",
			statusCode: fasthttp.StatusBadRequest,
			request: `{
				"city": null,
				"place": "River"
			}`,
			storeMethods: []StoreMethod{
				{
					method:     "GetLocation",
					args:       []interface{}{1, mock.AnythingOfType("*main.Location")},
					returnArgs: []interface{}{nil},
					run: func(args mock.Arguments) {
						location := args.Get(1).(*Location)
						*location = Location{
							ID:       1,
							Country:  "Russia",
							City:     "Moscow",
							Place:    "Some Place",
							Distance: 150,
						}
					},
				},
			},
		},
		{
			name:       "UpdateLocation/ErrUpdateID",
			path:       "/locations/1",
			statusCode: fasthttp.StatusBadRequest,
			request:    `{"id":2,"place": "another place"}`,
			storeMethods: []StoreMethod{
				{
					method:     "GetLocation",
					args:       []interface{}{1, mock.AnythingOfType("*main.Location")},
					returnArgs: []interface{}{nil},
				},
				{
					method:     "UpdateLocation",
					args:       []interface{}{1, mock.AnythingOfType("*main.Location")},
					returnArgs: []interface{}{ErrUpdateID},
				},
			},
		},
		{
			name:     "GetLocation",
			path:     "/locations/1",
			response: `{"id":1,"city":"Moscow","country":"Russia","place":"Some Place","distance":150}` + "\n",
			storeMethods: []StoreMethod{
				{
					method:     "GetLocation",
					args:       []interface{}{1, mock.AnythingOfType("*main.Location")},
					returnArgs: []interface{}{nil},
					run: func(args mock.Arguments) {
						location := args.Get(1).(*Location)
						*location = Location{
							ID:       1,
							Country:  "Russia",
							City:     "Moscow",
							Place:    "Some Place",
							Distance: 150,
						}
					},
				},
			},
		},
		{
			name:       "GetLocation/InvalidID",
			path:       "/locations/a",
			statusCode: fasthttp.StatusNotFound,
		},
		{
			name:       "GetLocation/NotFound",
			path:       "/locations/1",
			statusCode: fasthttp.StatusNotFound,
			storeMethods: []StoreMethod{
				{
					method:     "GetLocation",
					args:       []interface{}{1, mock.AnythingOfType("*main.Location")},
					returnArgs: []interface{}{mgo.ErrNotFound},
				},
			},
		},
		{
			name:     "GetLocationAvg",
			path:     "/locations/1/avg",
			response: `{"avg":4.375}` + "\n",
			storeMethods: []StoreMethod{
				{
					method:     "GetLocationAvg",
					args:       []interface{}{1, &LocationAvgQuery{}},
					returnArgs: []interface{}{4.375, nil},
				},
			},
		},
		{
			name:       "GetLocationAvg/InvalidID",
			path:       "/locations/a/avg",
			statusCode: fasthttp.StatusNotFound,
		},
		{
			name:       "GetLocationAvg/NotFound",
			path:       "/locations/999/avg",
			statusCode: fasthttp.StatusNotFound,
			storeMethods: []StoreMethod{
				{
					method:     "GetLocationAvg",
					args:       []interface{}{999, &LocationAvgQuery{}},
					returnArgs: []interface{}{0, mgo.ErrNotFound},
				},
			},
		},
		{
			name:     "GetLocationAvg/WithQuery",
			path:     "/locations/1/avg",
			query:    "?fromAge=30&toAge=40&gender=m",
			response: `{"avg":2.664}` + "\n",
			storeMethods: []StoreMethod{
				{
					method:     "GetLocationAvg",
					args:       []interface{}{1, &LocationAvgQuery{FromAge: 30, ToAge: 40, Gender: "m"}},
					returnArgs: []interface{}{2.664, nil},
				},
			},
		},
		{
			name:       "GetLocationAvg/WithInvalidQuery",
			path:       "/locations/1/avg",
			query:      "?toDate=a",
			statusCode: fasthttp.StatusBadRequest,
		},
		{
			name:     "GetLocationAvg/WithUnknownQuery",
			path:     "/locations/200/avg",
			query:    "?unknown=value",
			response: `{"avg":0}` + "\n",
			storeMethods: []StoreMethod{
				{
					method:     "GetLocationAvg",
					args:       []interface{}{200, &LocationAvgQuery{}},
					returnArgs: []interface{}{0, nil},
				},
			},
		},
		{
			name:     "GetLocationAvg/Rounding",
			path:     "/locations/15/avg",
			response: `{"avg":2.65217}` + "\n",
			storeMethods: []StoreMethod{
				{
					method:     "GetLocationAvg",
					args:       []interface{}{15, &LocationAvgQuery{}},
					returnArgs: []interface{}{2.652173913043478, nil},
				},
			},
		},
		{
			name:       "GetLocationAvg/ValidateQuery",
			path:       "/locations/15/avg",
			query:      "?gender=asd",
			statusCode: fasthttp.StatusBadRequest,
		},
		//-------------------------------
		// Visit endpoints tests
		//-------------------------------
		{
			name:     "CreateVisit",
			path:     "/visits/new",
			request:  `{"id":100,"user":1,"location":15,"visited_at":1268006400,"mark":5}`,
			response: "{}\n",
			storeMethods: []StoreMethod{
				{
					method:     "CreateVisit",
					args:       []interface{}{mock.AnythingOfType("*main.Visit")},
					returnArgs: []interface{}{nil},
				},
			},
		},
		{
			name:       "CreateVisit/InvalidBody",
			path:       "/visits/new",
			request:    `{bad-json}`,
			statusCode: fasthttp.StatusBadRequest,
		},
		{
			name:       "CreateVisit/ValidationError",
			path:       "/visits/new",
			request:    `{"mark":-10}`,
			statusCode: fasthttp.StatusBadRequest,
		},
		{
			name:       "CreateVisit/WithNullField",
			path:       "/visits/new",
			request:    `{"id":100,"user":1,"location":15,"visited_at":null,"mark":5}`,
			statusCode: fasthttp.StatusBadRequest,
		},
		{
			name:       "CreateVisit/DatabaseError",
			path:       "/visits/new",
			request:    `{"id":100,"user":1,"location":15,"visited_at":1268006400,"mark":5}`,
			statusCode: fasthttp.StatusInternalServerError,
			storeMethods: []StoreMethod{
				{
					method:     "CreateVisit",
					args:       []interface{}{mock.AnythingOfType("*main.Visit")},
					returnArgs: []interface{}{errors.New("db error")},
				},
			},
		},
		{
			name:     "UpdateVisit",
			path:     "/visits/100",
			request:  `{"mark":4}`,
			response: "{}\n",
			storeMethods: []StoreMethod{
				{
					method:     "GetVisit",
					args:       []interface{}{100, mock.AnythingOfType("*main.Visit")},
					returnArgs: []interface{}{nil},
					run: func(args mock.Arguments) {
						v := args.Get(1).(*Visit)
						v.LocationID = 15
					},
				},
				{
					method:     "UpdateVisit",
					args:       []interface{}{100, mock.AnythingOfType("*main.Visit")},
					returnArgs: []interface{}{nil},
					run: func(args mock.Arguments) {
						v := args.Get(1).(*Visit)
						assert.Equal(t, 4, v.Mark)
						assert.Equal(t, 15, v.LocationID)
					},
				},
			},
		},
		{
			name:       "UpdateVisit/InvalidID",
			path:       "/visits/a",
			statusCode: fasthttp.StatusNotFound,
		},
		{
			name:       "UpdateVisit/NotFound",
			path:       "/visits/998",
			statusCode: fasthttp.StatusNotFound,
			storeMethods: []StoreMethod{
				{
					method:     "GetVisit",
					args:       []interface{}{998, mock.AnythingOfType("*main.Visit")},
					returnArgs: []interface{}{mgo.ErrNotFound},
				},
			},
		},
		{
			name:       "UpdateVisit/InvalidBody",
			path:       "/visits/1",
			statusCode: fasthttp.StatusBadRequest,
			request:    `{bad-json}`,
			storeMethods: []StoreMethod{
				{
					method:     "GetVisit",
					args:       []interface{}{1, mock.AnythingOfType("*main.Visit")},
					returnArgs: []interface{}{nil},
				},
			},
		},
		{
			name:       "UpdateVisit/ValidationError",
			path:       "/visits/1",
			statusCode: fasthttp.StatusBadRequest,
			request:    `{"visited_at":900000000}`,
			storeMethods: []StoreMethod{
				{
					method:     "GetVisit",
					args:       []interface{}{1, mock.AnythingOfType("*main.Visit")},
					returnArgs: []interface{}{nil},
				},
			},
		},
		{
			name: "UpdateVisit/WithNullField",
			path: "/visits/1",
			request: `{
				"user": null,
				"location": 51530
			}`,
			statusCode: fasthttp.StatusBadRequest,
			storeMethods: []StoreMethod{
				{
					method:     "GetVisit",
					args:       []interface{}{1, mock.AnythingOfType("*main.Visit")},
					returnArgs: []interface{}{nil},
					run: func(args mock.Arguments) {
						visit := args.Get(1).(*Visit)
						*visit = Visit{
							ID:         99,
							UserID:     1,
							LocationID: 72,
							VisitedAt:  Timestamp{time.Unix(1268006400, 0)},
							Mark:       2,
						}
					},
				},
			},
		},
		{
			name:       "UpdateVisit/ErrUpdateID",
			path:       "/visits/1",
			statusCode: fasthttp.StatusBadRequest,
			request:    `{"id":2,"mark": 5}`,
			storeMethods: []StoreMethod{
				{
					method:     "GetVisit",
					args:       []interface{}{1, mock.AnythingOfType("*main.Visit")},
					returnArgs: []interface{}{nil},
				},
				{
					method:     "UpdateVisit",
					args:       []interface{}{1, mock.AnythingOfType("*main.Visit")},
					returnArgs: []interface{}{ErrUpdateID},
				},
			},
		},
		{
			name:     "GetVisit",
			path:     "/visits/99",
			response: `{"id":99,"user":1,"location":72,"visited_at":378654317,"mark":2}` + "\n",
			storeMethods: []StoreMethod{
				{
					method:     "GetVisit",
					args:       []interface{}{99, mock.AnythingOfType("*main.Visit")},
					returnArgs: []interface{}{nil},
					run: func(args mock.Arguments) {
						visit := args.Get(1).(*Visit)
						*visit = Visit{
							ID:         99,
							UserID:     1,
							LocationID: 72,
							VisitedAt:  Timestamp{time.Unix(378654317, 0)},
							Mark:       2,
						}
					},
				},
			},
		},
		{
			name:       "GetVisit/InvalidID",
			path:       "/visits/a",
			statusCode: fasthttp.StatusNotFound,
		},
		{
			name:       "GetVisit/NotFound",
			path:       "/visits/1",
			statusCode: fasthttp.StatusNotFound,
			storeMethods: []StoreMethod{
				{
					method:     "GetVisit",
					args:       []interface{}{1, mock.AnythingOfType("*main.Visit")},
					returnArgs: []interface{}{mgo.ErrNotFound},
				},
			},
		},
	}
	// Disable logging
	logrus.SetOutput(ioutil.Discard)
	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close()
	srv := NewServer(nil)
	go fasthttp.Serve(ln, srv.handler())
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// new store for each test
			store := new(MockStore)
			srv.store = store
			for _, sm := range tc.storeMethods {
				c := store.On(sm.method, sm.args...).Return(sm.returnArgs...)
				if sm.run != nil {
					c.Run(sm.run)
				}
			}

			req := fasthttp.AcquireRequest()
			req.SetRequestURI("http://localhost" + tc.path + tc.query)
			if tc.request != "" {
				req.Header.SetMethod("POST")
				req.SetBodyString(tc.request)
			}

			res := fasthttp.AcquireResponse()
			client := fasthttp.Client{
				Dial: func(_ string) (net.Conn, error) { return ln.Dial() },
			}

			if err := client.Do(req, res); err != nil {
				t.Fatalf("could not send request: %v", err)
			}

			statusCode := tc.statusCode
			if statusCode == 0 {
				statusCode = fasthttp.StatusOK
			}

			assert.Equal(t, statusCode, res.StatusCode(), "invalid status code")
			assert.Equal(t, tc.response, string(res.Body()), "invalid response body")
			if len(tc.response) > 0 {
				assert.Equal(t,
					"application/json; charset=utf-8",
					string(res.Header.Peek("Content-Type")),
					"invalid content type header")
			}
			store.AssertExpectations(t)

			fasthttp.ReleaseResponse(res)
			fasthttp.ReleaseRequest(req)
		})
	}
}
