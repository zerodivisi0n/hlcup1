package main

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/buger/jsonparser"

	"gopkg.in/mgo.v2/bson"
)

type User struct {
	ID        int       `json:"id" bson:"_id" validate:"omitempty,gt=0"`
	FirstName string    `json:"first_name" bson:"f" validate:"omitempty,max=50"`
	LastName  string    `json:"last_name" bson:"l" validate:"omitempty,max=50"`
	Email     string    `json:"email" bson:"e" validate:"omitempty,email,max=100"`
	Gender    string    `json:"gender" bson:"g" validate:"omitempty,eq=m|eq=f"`
	BirthDate Timestamp `json:"birth_date" bson:"b" validate:"omitempty,gte=-1262304000,lte=915148800"` // 01.01.1930 - 01.01.1999
}

type Location struct {
	ID       int    `json:"id" bson:"_id" validate:"omitempty,gt=0"`
	City     string `json:"city" bson:"ci" validate:"omitempty,max=50"`
	Country  string `json:"country" bson:"co" validate:"omitempty,max=50"`
	Place    string `json:"place" bson:"p"`
	Distance int    `json:"distance" bson:"d" validate:"omitempty,gte=0"`
}

type Visit struct {
	ID         int       `json:"id" bson:"_id" validate:"omitempty,gt=0"`
	UserID     int       `json:"user" bson:"u" validate:"omitempty,gt=0"`
	LocationID int       `json:"location" bson:"l" validate:"omitempty,gt=0"`
	VisitedAt  Timestamp `json:"visited_at" bson:"v" validate:"omitempty,gte=946684800,lte=1420070400"` // 01.01.2000 - 01.01.2015
	Mark       int       `json:"mark" bson:"m" validate:"omitempty,gte=0,lte=5"`
}

type UserVisit struct {
	Mark      int       `json:"mark" bson:"m"`
	VisitedAt Timestamp `json:"visited_at" bson:"v"`
	Place     string    `json:"place" bson:"p"`
}

type UserVisitsQuery struct {
	FromDate   Timestamp
	ToDate     Timestamp
	Country    string
	ToDistance int
}

type LocationAvgQuery struct {
	FromDate Timestamp
	ToDate   Timestamp
	FromAge  int
	ToAge    int
	Gender   string
}

type Timestamp struct {
	time.Time
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	ts := t.Time.Unix()
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	ts, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}

	t.SetUnix(int64(ts))

	return nil
}

func (t Timestamp) GetBSON() (interface{}, error) {
	if time.Time(t.Time).IsZero() {
		return nil, nil
	}

	return time.Time(t.Time), nil
}

func (t *Timestamp) SetBSON(raw bson.Raw) error {
	var tm time.Time

	if err := raw.Unmarshal(&tm); err != nil {
		return err
	}

	*t = Timestamp{tm}

	return nil
}

func (t *Timestamp) UnmarshalText(text []byte) error {
	return t.UnmarshalJSON(text)
}

func (t *Timestamp) String() string {
	return time.Time(t.Time).String()
}

func (t *Timestamp) SetUnix(ts int64) {
	t.Time = time.Unix(ts, 0)
}

func ValidateTimestamp(field reflect.Value) interface{} {
	if timestamp, ok := field.Interface().(Timestamp); ok {
		if !timestamp.Time.IsZero() {
			return timestamp.Time.Unix()
		}
	}
	return nil
}

// Custom unmarshalers
func (u *User) UnmarshalJSON(b []byte) error {
	return jsonparser.ObjectEach(b, func(key []byte, value []byte, vt jsonparser.ValueType, offset int) error {
		if vt == jsonparser.Null {
			return errors.New("null type")
		}
		if bytes.Equal(key, []byte("id")) {
			if id, err := jsonparser.ParseInt(value); err == nil {
				u.ID = int(id)
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
				u.BirthDate.SetUnix(ts)
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
				l.ID = int(id)
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
			if id, err := jsonparser.ParseInt(value); err == nil {
				l.Distance = int(id)
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
				v.ID = int(id)
			} else {
				return fmt.Errorf("invalid id: %v", err)
			}
		} else if bytes.Equal(key, []byte("user")) {
			if id, err := jsonparser.ParseInt(value); err == nil {
				v.UserID = int(id)
			} else {
				return fmt.Errorf("invalid user id: %v", err)
			}
		} else if bytes.Equal(key, []byte("location")) {
			if id, err := jsonparser.ParseInt(value); err == nil {
				v.LocationID = int(id)
			} else {
				return fmt.Errorf("invalid location id: %v", err)
			}
		} else if bytes.Equal(key, []byte("visited_at")) {
			if ts, err := jsonparser.ParseInt(value); err == nil {
				v.VisitedAt.SetUnix(ts)
			} else {
				return fmt.Errorf("invalid visited_at: %v", err)
			}
		} else if bytes.Equal(key, []byte("mark")) {
			if mark, err := jsonparser.ParseInt(value); err == nil {
				v.Mark = int(mark)
			} else {
				return fmt.Errorf("invalid mark: %v", err)
			}
		}
		return nil
	})
}
