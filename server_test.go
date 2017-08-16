package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandlers(t *testing.T) {
	srv := NewServer(nil)
	type StoreMethod struct {
		method     string
		args       []interface{}
		returnArgs []interface{}
		run        func(args mock.Arguments)
	}
	tt := []struct {
		name         string
		handler      httprouter.Handle
		entityID     string
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
			handler:  srv.updateUser,
			entityID: "new",
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
			handler:    srv.updateUser,
			entityID:   "new",
			request:    `{bad-json}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "CreateUser/WithoutID",
			handler:    srv.updateUser,
			entityID:   "new",
			request:    `{"first_name":"User"}`,
			statusCode: http.StatusBadRequest,
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
			handler:    srv.updateUser,
			entityID:   "new",
			request:    `{"first_name":"Alien","gender":"u"}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "CreateUser/DatabaseError",
			handler:    srv.updateUser,
			entityID:   "new",
			request:    `{"id":1,"first_name":"First","last_name":"User","email":"foo@bar.com","gender":"m","birth_date":100000}`,
			statusCode: http.StatusInternalServerError,
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
			handler:    srv.updateUser,
			entityID:   "new",
			request:    `{"id":1,"first_name":"First","last_name":"User","email":"duplicate@email.com","gender":"m","birth_date":100000}`,
			statusCode: http.StatusBadRequest,
			storeMethods: []StoreMethod{
				{
					method:     "CreateUser",
					args:       []interface{}{mock.AnythingOfType("*main.User")},
					returnArgs: []interface{}{&mgo.LastError{Code: 11000}},
				},
			},
		},
		{
			name:     "UpdateUser",
			handler:  srv.updateUser,
			entityID: "1",
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
			handler:    srv.updateUser,
			entityID:   "a",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "UpdateUser/NotFound",
			handler:    srv.updateUser,
			entityID:   "2",
			statusCode: http.StatusNotFound,
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
			handler:    srv.updateUser,
			entityID:   "1",
			statusCode: http.StatusBadRequest,
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
			handler:    srv.updateUser,
			entityID:   "1",
			statusCode: http.StatusBadRequest,
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
			handler:    srv.updateUser,
			entityID:   "1",
			statusCode: http.StatusBadRequest,
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
			name:     "GetUser",
			handler:  srv.getUser,
			entityID: "1",
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
			handler:    srv.getUser,
			entityID:   "a",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "GetUser/NotFound",
			handler:    srv.getUser,
			entityID:   "1",
			statusCode: http.StatusNotFound,
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
			handler:  srv.getUserVisits,
			entityID: "1",
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
			handler:    srv.getUserVisits,
			entityID:   "a",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "GetUserVisits/NotFound",
			handler:    srv.getUserVisits,
			entityID:   "999",
			statusCode: http.StatusNotFound,
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
			handler:  srv.getUserVisits,
			entityID: "2",
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
			handler:  srv.getUserVisits,
			entityID: "1",
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
			handler:    srv.getUserVisits,
			entityID:   "1",
			query:      "?toDate=a",
			statusCode: http.StatusBadRequest,
		},
		{
			name:     "GetUserVisits/WithUnknownQuery",
			handler:  srv.getUserVisits,
			entityID: "1",
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
			handler:  srv.updateLocation,
			entityID: "new",
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
			handler:    srv.updateLocation,
			entityID:   "new",
			request:    `{bad-json}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "CreateLocation/ValidationError",
			handler:    srv.updateLocation,
			entityID:   "new",
			request:    `{"distance":-5}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "CreateUser/DatabaseError",
			handler:    srv.updateLocation,
			entityID:   "new",
			request:    `{"id":1,"city":"Moscow","country":"Russia","place":"Some Place","distance":25}`,
			statusCode: http.StatusInternalServerError,
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
			handler:  srv.updateLocation,
			entityID: "1",
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
			handler:    srv.updateLocation,
			entityID:   "a",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "UpdateLocation/NotFound",
			handler:    srv.updateLocation,
			entityID:   "2",
			statusCode: http.StatusNotFound,
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
			handler:    srv.updateLocation,
			entityID:   "1",
			statusCode: http.StatusBadRequest,
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
			handler:    srv.updateLocation,
			entityID:   "1",
			statusCode: http.StatusBadRequest,
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
			name:       "UpdateLocation/ErrUpdateID",
			handler:    srv.updateLocation,
			entityID:   "1",
			statusCode: http.StatusBadRequest,
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
			handler:  srv.getLocation,
			entityID: "1",
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
			handler:    srv.getLocation,
			entityID:   "a",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "GetLocation/NotFound",
			handler:    srv.getLocation,
			entityID:   "1",
			statusCode: http.StatusNotFound,
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
			handler:  srv.getLocationAvg,
			entityID: "1",
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
			handler:    srv.getLocationAvg,
			entityID:   "a",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "GetLocationAvg/NotFound",
			handler:    srv.getLocationAvg,
			entityID:   "999",
			statusCode: http.StatusNotFound,
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
			handler:  srv.getLocationAvg,
			entityID: "1",
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
			handler:    srv.getLocationAvg,
			entityID:   "1",
			query:      "?toDate=a",
			statusCode: http.StatusBadRequest,
		},
		{
			name:     "GetLocationAvg/WithUnknownQuery",
			handler:  srv.getLocationAvg,
			entityID: "200",
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
			handler:  srv.getLocationAvg,
			entityID: "15",
			response: `{"avg":2.65217}` + "\n",
			storeMethods: []StoreMethod{
				{
					method:     "GetLocationAvg",
					args:       []interface{}{15, &LocationAvgQuery{}},
					returnArgs: []interface{}{2.652173913043478, nil},
				},
			},
		},
		//-------------------------------
		// Visit endpoints tests
		//-------------------------------
		{
			name:     "CreateVisit",
			handler:  srv.updateVisit,
			entityID: "new",
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
			handler:    srv.updateVisit,
			entityID:   "new",
			request:    `{bad-json}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "CreateVisit/ValidationError",
			handler:    srv.updateVisit,
			entityID:   "new",
			request:    `{"mark":-10}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "CreateVisit/DatabaseError",
			handler:    srv.updateVisit,
			entityID:   "new",
			request:    `{"id":100,"user":1,"location":15,"visited_at":1268006400,"mark":5}`,
			statusCode: http.StatusInternalServerError,
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
			handler:  srv.updateVisit,
			entityID: "100",
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
			handler:    srv.updateVisit,
			entityID:   "a",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "UpdateVisit/NotFound",
			handler:    srv.updateVisit,
			entityID:   "998",
			statusCode: http.StatusNotFound,
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
			handler:    srv.updateVisit,
			entityID:   "1",
			statusCode: http.StatusBadRequest,
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
			handler:    srv.updateVisit,
			entityID:   "1",
			statusCode: http.StatusBadRequest,
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
			name:       "UpdateVisit/ErrUpdateID",
			handler:    srv.updateVisit,
			entityID:   "1",
			statusCode: http.StatusBadRequest,
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
			handler:  srv.getVisit,
			entityID: "99",
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
			handler:    srv.getVisit,
			entityID:   "a",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "GetVisit/NotFound",
			handler:    srv.getVisit,
			entityID:   "1",
			statusCode: http.StatusNotFound,
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
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader
			if tc.request != "" {
				body = strings.NewReader(tc.request)
			}
			req, err := http.NewRequest("GET", "/test/request"+tc.query, body)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}
			rec := httptest.NewRecorder()
			var params httprouter.Params
			if tc.entityID != "" {
				params = httprouter.Params{
					httprouter.Param{Key: "id", Value: tc.entityID},
				}
			}

			// new store for each test
			store := new(MockStore)
			srv.store = store
			for _, sm := range tc.storeMethods {
				c := store.On(sm.method, sm.args...).Return(sm.returnArgs...)
				if sm.run != nil {
					c.Run(sm.run)
				}
			}

			tc.handler(rec, req, params)

			res := rec.Result()
			defer res.Body.Close()

			buf, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("could not read response: %v", err)
			}

			statusCode := tc.statusCode
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			assert.Equal(t, statusCode, res.StatusCode, "invalid status code")
			assert.Equal(t, tc.response, string(buf), "invalid response body")
			if len(tc.response) > 0 {
				assert.Equal(t,
					"application/json; charset=utf-8",
					res.Header.Get("Content-Type"),
					"invalid content type header")
			}
			store.AssertExpectations(t)
		})
	}
}

func TestRouting(t *testing.T) {
	store := new(MockStore)
	srv := httptest.NewServer(NewServer(store).handler())
	defer srv.Close()

	store.On("GetUser", 1, mock.AnythingOfType("*main.User")).Return(nil).Run(func(args mock.Arguments) {
		user := args.Get(1).(*User)
		*user = User{
			ID:        1,
			FirstName: "First",
			LastName:  "User",
			Email:     "foo@bar.com",
			Gender:    "m",
			BirthDate: Timestamp{time.Unix(100000, 0)},
		}
	})

	res, err := http.Get(fmt.Sprintf("%s/users/1", srv.URL))
	if err != nil {
		t.Fatalf("could not send GET request: %v", err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode, "invalid status code")
	assert.Equal(t,
		"application/json; charset=utf-8",
		res.Header.Get("Content-Type"),
		"invalid content type header")

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("could not read response: %v", err)
	}
	assert.Equal(t,
		string(buf),
		`{"id":1,"first_name":"First","last_name":"User","email":"foo@bar.com","gender":"m","birth_date":100000}`+"\n",
		"invalid response")
}
