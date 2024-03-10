package version

import (
	"strconv"
	"strings"
)

var (
	BuildVersion string = ""
	BuildTime    string = ""
)

type Version struct {
	Major int
	Minor int
	Build int
}

func New(s string) *Version {
	v := &Version{}
	numbers := strings.Split(s, ".")
	if len(numbers) == 0 {
		return v
	}
	if len(numbers) > 0 {
		i, err := strconv.Atoi(numbers[0])
		if err != nil {
			return v
		}
		v.Major = i
	}
	if len(numbers) > 1 {
		i, err := strconv.Atoi(numbers[1])
		if err != nil {
			return v
		}
		v.Minor = i
	}
	if len(numbers) > 2 {
		i, err := strconv.Atoi(numbers[2])
		if err != nil {
			return v
		}
		v.Build = i
	}
	return v
}

func (v *Version) IsEqual(v2 *Version) bool {
	return v.Major == v2.Major && v.Minor == v2.Minor && v.Build == v2.Build
}

func (v *Version) IsSupported(versions ...*Version) bool {
	for _, supported := range versions {
		if v.IsEqual(supported) {
			return true
		}
	}
	return false
}
