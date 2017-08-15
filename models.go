package main

import (
	"fmt"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type User struct {
	ID        int       `json:"id" bson:"_id"`
	FirstName string    `json:"first_name" bson:"f"`
	LastName  string    `json:"last_name" bson:"l"`
	Email     string    `json:"email" bson:"e"`
	Gender    string    `json:"gender" bson:"g"`
	BirthDate Timestamp `json:"birth_date" bson:"b"`
}

type Location struct {
	ID       int    `json:"id" bson:"_id"`
	City     string `json:"city" bson:"ci"`
	Country  string `json:"country" bson:"co"`
	Place    string `json:"place" bson:"p"`
	Distance int    `json:"distance" bson:"d"`
}

type Visit struct {
	ID         int       `json:"id" bson:"_id"`
	UserID     int       `json:"user" bson:"u"`
	LocationID int       `json:"location" bson:"l"`
	VisitedAt  Timestamp `json:"visited_at" bson:"v"`
	Mark       int       `json:"mark" bson:"m"`
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

	t.Time = time.Unix(int64(ts), 0)

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

func (t *Timestamp) String() string {
	return time.Time(t.Time).String()
}
