package models

import (
	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/listing"
	"github.com/themisir/myfeed/pkg/removing"
	"github.com/themisir/myfeed/pkg/updating"
)

type SourceRepository interface {
	adding.SourceRepository
	listing.SourceRepository
	removing.SourceRepository
	updating.SourceRepository
}
