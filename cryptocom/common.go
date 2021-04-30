package cryptocom

import (
	"errors"
	"strings"
)

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
