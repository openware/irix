package cryptocom

import (
	"errors"
	"github.com/openware/pkg/validate"
	"strings"
)

func tryOrError(checks ...validate.Check) (err error) {
	for _, fn := range checks {
		if err = fn(); err != nil {
			return err
		}
	}
	return
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
