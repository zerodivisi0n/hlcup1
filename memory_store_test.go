package main

import "testing"
import "github.com/stretchr/testify/assert"

func TestUsers(t *testing.T) {
	s := NewMemoryStore()
	u1 := User{ID: 1, Email: "foo@bar.com"}

	// test create
	err := s.CreateUser(&u1)
	assert.NoError(t, err)
	err = s.CreateUsers([]User{
		User{ID: 2, Email: "user2@hlcup.com", FirstName: "User2"},
		User{ID: 3, Email: "user3@hlcup.com", FirstName: "User3"},
		User{ID: 4, Email: "user4@hlcup.com", FirstName: "User4"},
		User{ID: 5, Email: "user5@hlcup.com", FirstName: "User6"},
	})
	assert.NoError(t, err)

	// test get
	var u User
	err = s.GetUser(1, &u)
	assert.NoError(t, err)
	assert.Equal(t, &u1, &u)

	// test update
	u.Email = "updated@user.com"
	err = s.UpdateUser(1, &u)
	assert.NoError(t, err)
	err = s.GetUser(1, &u)
	assert.NoError(t, err)
	assert.Equal(t, "updated@user.com", u.Email)
}

func TestLocations(t *testing.T) {

}

func TestVisits(t *testing.T) {

}

func TestUpdateVisit(t *testing.T) {
	u1 := User{ID: 1, FirstName: "User1", Email: "foo@bar.com"}
	u2 := User{ID: 2, FirstName: "User2", Email: "foo@baz.com"}

	l1 := Location{ID: 1, Place: "Place1"}
	l2 := Location{ID: 2, Place: "Place2"}
	l3 := Location{ID: 3, Place: "Place3"}

	v1 := Visit{ID: 1, UserID: 1, LocationID: 1, VisitedAt: &[]int64{100}[0], Mark: &[]int{2}[0]}
	v2 := Visit{ID: 2, UserID: 2, LocationID: 2, VisitedAt: &[]int64{200}[0], Mark: &[]int{3}[0]}
	v3 := Visit{ID: 3, UserID: 1, LocationID: 3, VisitedAt: &[]int64{300}[0], Mark: &[]int{4}[0]}

	s := NewMemoryStore()
	assert.NoError(t, s.CreateUser(&u1))
	assert.NoError(t, s.CreateUser(&u2))
	assert.NoError(t, s.CreateLocation(&l1))
	assert.NoError(t, s.CreateLocation(&l2))
	assert.NoError(t, s.CreateLocation(&l3))
	assert.NoError(t, s.CreateVisit(&v1))
	assert.NoError(t, s.CreateVisit(&v2))
	assert.NoError(t, s.CreateVisit(&v3))

	v3u := Visit{ID: 3, UserID: 2, LocationID: 3, VisitedAt: &[]int64{300}[0], Mark: &[]int{2}[0]}
	assert.NoError(t, s.UpdateVisit(3, &v3u))

	v2u := Visit{ID: 2, UserID: 2, LocationID: 1, VisitedAt: &[]int64{150}[0], Mark: &[]int{2}[0]}
	assert.NoError(t, s.updateVisit(2, &v2u))
}
