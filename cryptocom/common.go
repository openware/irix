package cryptocom

import (
	"errors"
	"github.com/openware/pkg/validate"
	"regexp"
	"strings"
	"time"
)

func tryOrError(checks ...validate.Check) (err error) {
	for _, fn := range checks {
		if err = fn(); err != nil {
			return err
		}
	}
	return
}
func timestampMs(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func validInstrument(instrument string) (err error) {
	splits := strings.Split(instrument, "_")
	if instrument == "" ||
		len(splits) != 2 ||
		splits[0] == "" ||
		splits[1] == "" {
		err = errors.New("invalid instrument name value")
		return
	}
	return
}

func validPagination(pageSize, page int) error {
	if pageSize < 0 {
		return errors.New("page size should be at least 0")
	}
	if pageSize > 200 {
		return errors.New("max page size is 200")
	}
	if page < 0 {
		return errors.New("page should be at least 0")
	}
	return nil
}

func isValidCurrency(code string) (err error) {
	regex := regexp.MustCompile("^[a-zA-Z0-9]+$")
	if code == "" || len(code) < 3 || !regex.MatchString(code) {
		err = errors.New("invalid code")
	}
	return
}

