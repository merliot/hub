package nmea

import (
	"errors"

	_nmea "github.com/adrianmo/go-nmea"
)

func ParseGLL(text string) (lat, long float64, err error) {
	var rec _nmea.Sentence

	rec, err = _nmea.Parse(text)
	if err != nil {
		return
	}
	if rec.DataType() != _nmea.TypeGLL {
		err = errors.New("DataType not GLL")
		return
	}
	gll := rec.(_nmea.GLL)
	if gll.Validity != "A" {
		err = errors.New("GLL not valid")
		return
	}

	lat, long = gll.Latitude, gll.Longitude
	return
}
