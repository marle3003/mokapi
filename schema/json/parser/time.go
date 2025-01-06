package parser

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	parsingUnit  = iota
	parsingValue = 1
)

func ParseDuration(s string) error {
	state := parsingUnit
	num := ""
	var err error

	switch {
	case strings.HasPrefix(s, "P"):
	case strings.HasPrefix(s, "-P"):
		s = s[1:]
	default:
		return fmt.Errorf("invalid duration format: %s", s)
	}

	for _, char := range s {
		switch char {
		case 'P':
			if state != parsingUnit {
				return fmt.Errorf("invalid duration format: %s", s)
			}
		case 'T':
			state = parsingValue
		case 'Y':
			if state != parsingUnit {
				return fmt.Errorf("invalid duration format: %s", s)
			}

			_, err = strconv.ParseFloat(num, 64)
			if err != nil {
				return fmt.Errorf("invalid duration format: %s", s)
			}
			num = ""
		case 'M':
			if state == parsingUnit {
				_, err = strconv.ParseFloat(num, 64)
				if err != nil {
					return fmt.Errorf("invalid duration format: %s", s)
				}
				num = ""
			} else {
				_, err = strconv.ParseFloat(num, 64)
				if err != nil {
					return err
				}
				num = ""
			}
		case 'W':
			if state != parsingUnit {
				return fmt.Errorf("invalid duration format: %s", s)
			}

			_, err = strconv.ParseFloat(num, 64)
			if err != nil {
				return err
			}
			num = ""
		case 'D':
			if state != parsingUnit {
				return fmt.Errorf("invalid duration format: %s", s)
			}

			_, err = strconv.ParseFloat(num, 64)
			if err != nil {
				return err
			}
			num = ""
		case 'H':
			if state != parsingUnit {
				return fmt.Errorf("invalid duration format: %s", s)
			}

			_, err = strconv.ParseFloat(num, 64)
			if err != nil {
				return err
			}
			num = ""
		case 'S':
			if state != parsingUnit {
				return fmt.Errorf("invalid duration format: %s", s)
			}

			_, err = strconv.ParseFloat(num, 64)
			if err != nil {
				return err
			}
			num = ""
		default:
			if unicode.IsNumber(char) || char == '.' {
				num += string(char)
				continue
			}

			return fmt.Errorf("invalid duration format: %s", s)
		}
	}

	return nil
}
