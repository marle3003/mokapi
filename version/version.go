package version

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	BuildVersion = "0.0.0"
	BuildTime    = ""
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func New(s string) Version {
	v := Version{}
	v.parse(s)
	return v
}

func (v *Version) IsEqual(v2 Version) bool {
	return v.Major == v2.Major && v.Minor == v2.Minor && v.Patch == v2.Patch
}

func (v *Version) IsLower(v2 Version) bool {
	return v.Major < v2.Major || v.Minor < v2.Minor || v.Patch < v2.Patch
}

func (v *Version) IsSupported(versions ...Version) bool {
	for _, supported := range versions {
		if v.IsEqual(supported) {
			return true
		}
	}
	return false
}

func (v *Version) IsEmpty() bool {
	return v.Major == 0
}

func (v *Version) String() string {
	return fmt.Sprintf("%v.%v.%v", v.Major, v.Minor, v.Patch)
}

func (v *Version) parse(s string) {
	numbers := strings.Split(s, ".")
	i, err := strconv.Atoi(numbers[0])
	if err != nil {
		return
	}
	v.Major = i
	if len(numbers) > 1 {
		i, err := strconv.Atoi(numbers[1])
		if err != nil {
			return
		}
		v.Minor = i
	}
	if len(numbers) > 2 {
		i, err := strconv.Atoi(numbers[2])
		if err != nil {
			return
		}
		v.Patch = i
	}
}

func (v *Version) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	v.parse(s)
	return nil
}

func (v *Version) UnmarshalYAML(value *yaml.Node) error {
	var s string
	err := value.Decode(&s)
	if err != nil {
		return err
	}
	v.parse(s)
	return nil
}
