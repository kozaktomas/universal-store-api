package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemStorage(t *testing.T) {
	firstServiceName := "first_service"
	secondServiceName := "second_service"
	s := CreateMemStorage([]string{firstServiceName, secondServiceName})

	// add 3 items first service
	var err error
	_, err = s.Add(firstServiceName, 1)
	assert.Nil(t, err)
	_, err = s.Add(firstServiceName, 2)
	assert.Nil(t, err)
	_, err = s.Add(firstServiceName, 3)
	assert.Nil(t, err)

	// add 2 items second service
	_, err = s.Add(secondServiceName, 1)
	assert.Nil(t, err)
	_, err = s.Add(secondServiceName, 2)
	assert.Nil(t, err)

	// get list and check length
	list, err := s.List(firstServiceName)
	assert.Nil(t, err)
	assert.Len(t, list, 3)

	// test all items
	assert.Equal(t, list[0].Payload, 1)
	assert.Equal(t, list[1].Payload, 2)
	assert.Equal(t, list[2].Payload, 3)

	// delete 2 from list
	assert.Nil(t, s.Delete(firstServiceName, list[1].Id))
	assert.Equal(t, list[0].Payload, 1)
	assert.Equal(t, list[1].Payload, 3)
}
