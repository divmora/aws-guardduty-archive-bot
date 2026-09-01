package utils

import (
	"fmt"
	"strings"
)

// Contains checks if a string is present in a string slice.
func Contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// Arn represents the components of an Amazon Resource Name.
type Arn struct {
	Partition string
	Service   string
	Region    string
	AccountID string
	Resource  string
}

// ParseArn parses an AWS ARN string into its components.
func ParseArn(arn string) (*Arn, error) {
	if !strings.HasPrefix(arn, "arn:") {
		return nil, fmt.Errorf("invalid ARN: must start with 'arn:'")
	}

	parts := strings.SplitN(arn, ":", 6)
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid ARN format: %s", arn)
	}

	return &Arn{
		Partition: parts[1],
		Service:   parts[2],
		Region:    parts[3],
		AccountID: parts[4],
		Resource:  parts[5],
	}, nil
}
