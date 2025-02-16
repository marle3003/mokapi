package directory

import (
	"fmt"
	"mokapi/ldap"
)

type EntryError struct {
	Message string
	Code    uint8
}

type Entry struct {
	Dn         string
	Attributes map[string][]string
}

func (e *Entry) copy() Entry {
	c := Entry{
		Dn:         e.Dn,
		Attributes: make(map[string][]string),
	}
	for k, v := range e.Attributes {
		c.Attributes[k] = v
	}
	return c
}

func (e *Entry) validate(s *Schema) error {
	if s == nil {
		return nil
	}

	for name, values := range e.Attributes {
		if a, ok := s.AttributeTypes[name]; ok {
			for _, value := range values {
				if !a.Validate(value) {
					return NewEntryError(ldap.ConstraintViolation, "invalid value '%v' for attribute '%v': does not conform to required syntax (%v)", value, name, a.Syntax)
				}
			}
		}
	}

	classes, ok := e.Attributes["objectClass"]
	if !ok {
		return nil
	}
	for _, class := range classes {
		if c, ok := s.ObjectClasses[class]; ok {
			err := e.validateObjectClass(class, c, s)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *Entry) validateObjectClass(className string, c *ObjectClass, s *Schema) error {
	for _, name := range c.Must {
		if _, found := e.Attributes[name]; !found {
			return NewEntryError(ldap.ConstraintViolation, "entry is missing required attribute '%v': object class '%v' requires the following attributes: %v", name, className, c.Must)
		}
	}
	for _, name := range c.SuperClass {
		if super, ok := s.ObjectClasses[name]; ok {
			if err := e.validateObjectClass(name, super, s); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *EntryError) Error() string {
	return e.Message
}

func NewEntryError(code uint8, format string, args ...interface{}) *EntryError {
	return &EntryError{
		Message: fmt.Sprintf(format, args...),
		Code:    code,
	}
}
