package main

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/buger/jsonparser"
)

//easyjson:json
type User struct {
	ID        uint   `json:"id" bson:"_id"`
	FirstName string `json:"first_name" bson:"f"`
	LastName  string `json:"last_name" bson:"l"`
	Email     string `json:"email" bson:"e"`
	Gender    string `json:"gender" bson:"g"`
	BirthDate *int64 `json:"birth_date" bson:"b"`
	JSON      []byte `json:"-" bson:"j"`
}

//easyjson:json
type Location struct {
	ID       uint   `json:"id" bson:"_id"`
	City     string `json:"city" bson:"ci"`
	Country  string `json:"country" bson:"co"`
	Place    string `json:"place" bson:"p"`
	Distance *int   `json:"distance" bson:"d"`
	JSON     []byte `json:"-" bson:"j"`
}

//easyjson:json
type Visit struct {
	ID         uint   `json:"id" bson:"_id"`
	UserID     uint   `json:"user" bson:"u"`
	LocationID uint   `json:"location" bson:"l"`
	VisitedAt  *int64 `json:"visited_at" bson:"v"`
	Mark       *int   `json:"mark" bson:"m"`
	JSON       []byte `json:"-" bson:"j"`
}

//easyjson:json
type UserVisit struct {
	Mark      int    `json:"mark" bson:"m"`
	VisitedAt int64  `json:"visited_at" bson:"v"`
	Place     string `json:"place" bson:"p"`
}

type UserVisitsQuery struct {
	FromDate   *int64
	ToDate     *int64
	Country    string
	ToDistance *int
}

type LocationAvgQuery struct {
	FromDate *int64
	ToDate   *int64
	FromAge  *int
	ToAge    *int
	Gender   string
}

func (q LocationAvgQuery) FromBirth() *int64 {
	if q.ToAge == nil {
		return nil
	}
	from := time.Now().AddDate(-*q.ToAge, 0, 0).Unix()
	return &from
}

func (q LocationAvgQuery) ToBirth() *int64 {
	if q.FromAge == nil {
		return nil
	}
	to := time.Now().AddDate(-*q.FromAge, 0, 0).Unix()
	return &to
}

//easyjson:json
type UserVisitsResult struct {
	Visits []UserVisit `json:"visits"`
}

//easyjson:json
type LocationAvgResult struct {
	Avg float64 `json:"avg"`
}

// Custom unmarshalers
func (u *User) UnmarshalJSON(b []byte) error {
	return jsonparser.ObjectEach(b, func(key []byte, value []byte, vt jsonparser.ValueType, offset int) error {
		if vt == jsonparser.Null {
			return errors.New("null type")
		}
		if bytes.Equal(key, []byte("id")) {
			if id, err := jsonparser.ParseInt(value); err == nil {
				u.ID = uint(id)
			} else {
				return fmt.Errorf("invalid id: %v", err)
			}
		} else if bytes.Equal(key, []byte("first_name")) {
			if s, err := jsonparser.ParseString(value); err == nil {
				u.FirstName = s
			} else {
				return fmt.Errorf("invalid first name: %v", err)
			}
		} else if bytes.Equal(key, []byte("last_name")) {
			if s, err := jsonparser.ParseString(value); err == nil {
				u.LastName = s
			} else {
				return fmt.Errorf("invalid last name: %v", err)
			}
		} else if bytes.Equal(key, []byte("email")) {
			if s, err := jsonparser.ParseString(value); err == nil {
				u.Email = s
			} else {
				return fmt.Errorf("invalid email: %v", err)
			}
		} else if bytes.Equal(key, []byte("gender")) {
			if s, err := jsonparser.ParseString(value); err == nil {
				u.Gender = s
			} else {
				return fmt.Errorf("invalid gender: %v", err)
			}
		} else if bytes.Equal(key, []byte("birth_date")) {
			if ts, err := jsonparser.ParseInt(value); err == nil {
				u.BirthDate = &ts
			} else {
				return errors.New("invalid birth date")
			}
		}
		return nil
	})
}

func (l *Location) UnmarshalJSON(b []byte) error {
	return jsonparser.ObjectEach(b, func(key []byte, value []byte, vt jsonparser.ValueType, offset int) error {
		if vt == jsonparser.Null {
			return errors.New("null type")
		}
		if bytes.Equal(key, []byte("id")) {
			if id, err := jsonparser.ParseInt(value); err == nil {
				l.ID = uint(id)
			} else {
				return fmt.Errorf("invalid id: %v", err)
			}
		} else if bytes.Equal(key, []byte("city")) {
			if s, err := jsonparser.ParseString(value); err == nil {
				l.City = s
			} else {
				return fmt.Errorf("invalid city: %v", err)
			}
		} else if bytes.Equal(key, []byte("country")) {
			if s, err := jsonparser.ParseString(value); err == nil {
				l.Country = s
			} else {
				return fmt.Errorf("invalid country: %v", err)
			}
		} else if bytes.Equal(key, []byte("place")) {
			if s, err := jsonparser.ParseString(value); err == nil {
				l.Place = s
			} else {
				return fmt.Errorf("invalid place: %v", err)
			}
		} else if bytes.Equal(key, []byte("distance")) {
			if d, err := jsonparser.ParseInt(value); err == nil {
				d2 := int(d)
				l.Distance = &d2
			} else {
				return fmt.Errorf("invalid distance: %v", err)
			}
		}
		return nil
	})
}

func (v *Visit) UnmarshalJSON(b []byte) error {
	return jsonparser.ObjectEach(b, func(key []byte, value []byte, vt jsonparser.ValueType, offset int) error {
		if vt == jsonparser.Null {
			return errors.New("null type")
		}
		if bytes.Equal(key, []byte("id")) {
			if id, err := jsonparser.ParseInt(value); err == nil {
				v.ID = uint(id)
			} else {
				return fmt.Errorf("invalid id: %v", err)
			}
		} else if bytes.Equal(key, []byte("user")) {
			if id, err := jsonparser.ParseInt(value); err == nil {
				v.UserID = uint(id)
			} else {
				return fmt.Errorf("invalid user id: %v", err)
			}
		} else if bytes.Equal(key, []byte("location")) {
			if id, err := jsonparser.ParseInt(value); err == nil {
				v.LocationID = uint(id)
			} else {
				return fmt.Errorf("invalid location id: %v", err)
			}
		} else if bytes.Equal(key, []byte("visited_at")) {
			if ts, err := jsonparser.ParseInt(value); err == nil {
				v.VisitedAt = &ts
			} else {
				return fmt.Errorf("invalid visited_at: %v", err)
			}
		} else if bytes.Equal(key, []byte("mark")) {
			if mark, err := jsonparser.ParseInt(value); err == nil {
				mark2 := int(mark)
				v.Mark = &mark2
			} else {
				return fmt.Errorf("invalid mark: %v", err)
			}
		}
		return nil
	})
}

// Validators
func (u User) Validate() bool {
	return u.ID > 0 &&
		len(u.Email) > 0 && len(u.Email) < 100 &&
		len(u.FirstName) > 0 && len(u.LastName) < 50 &&
		len(u.LastName) > 0 && len(u.LastName) < 50 &&
		(u.Gender == "m" || u.Gender == "f") &&
		u.BirthDate != nil
}

func (l Location) Validate() bool {
	return l.ID > 0 &&
		len(l.Place) > 0 &&
		len(l.Country) > 0 && len(l.Country) < 50 &&
		len(l.City) > 0 && len(l.City) < 50 &&
		l.Distance != nil && *l.Distance > 0
}

func (v Visit) Validate() bool {
	return v.ID > 0 &&
		v.LocationID > 0 &&
		v.UserID > 0 &&
		v.VisitedAt != nil &&
		v.Mark != nil && *v.Mark >= 0 && *v.Mark <= 5
}
