package models

import (
	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/listing"
)

type UserRepository interface {
	adding.UserRepository
	listing.UserRepository
}
