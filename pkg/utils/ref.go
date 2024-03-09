package utils

import (
	"fmt"
	"strings"
)

type Ref string

func (r Ref) Parse() (namespace string, name string) {
	parts := strings.SplitN(string(r), "/", 2)
	if len(parts) == 1 {
		return "", parts[0]
	}
	return parts[0], parts[1]
}

func NewRef(namespace string, name string) string {
	if namespace == "" {
		return name
	}
	return fmt.Sprintf("%s/%s", namespace, name)
}
