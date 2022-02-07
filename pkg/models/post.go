package models

import (
	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/listing"
	"github.com/themisir/myfeed/pkg/removing"
	"github.com/themisir/myfeed/pkg/updating"
)

type PostRepository interface {
	adding.PostRepository
	listing.PostRepository
	removing.PostRepository
	updating.PostRepository
}
