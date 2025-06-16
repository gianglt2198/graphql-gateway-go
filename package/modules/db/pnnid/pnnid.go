package pnnid

import (
	"database/sql/driver"
	"fmt"
	"io"
	"strconv"

	"github.com/gianglt2198/federation-go/package/utils"
)

// ID implements a PNNID - a prefixed ULID.
type ID string

// MustNew returns a new PNNID for given a prefix
func MustNew(prefix string) ID { return ID(utils.NewID(21, prefix)) }

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (u *ID) UnmarshalGQL(v interface{}) error {
	return u.Scan(v)
}

// MarshalGQL implements the graphql.Marshaler interface
func (u ID) MarshalGQL(w io.Writer) {
	_, _ = io.WriteString(w, strconv.Quote(string(u)))
}

// Scan implements the Scanner interface.
func (u *ID) Scan(src interface{}) error {
	if src == nil {
		return fmt.Errorf("PNNID: expected a value")
	}
	switch src := src.(type) {
	case string:
		*u = ID(src)
	case ID:
		*u = src
	default:
		return fmt.Errorf("PNNID: unexpected type, %T", src)
	}
	return nil
}

// Value implements the driver Valuer interface.
func (u ID) Value() (driver.Value, error) {
	return string(u), nil
}
