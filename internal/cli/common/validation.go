package common

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateRequired returns an error if value is empty.
// Use for required command arguments.
//
// Example:
//
//	if err := common.ValidateRequired("event ID", args[0]); err != nil {
//	    return err
//	}
func ValidateRequired(name, value string) error {
	if value == "" {
		return NewUserError(
			fmt.Sprintf("%s is required", name),
			fmt.Sprintf("Provide %s as an argument", name),
		)
	}
	return nil
}

// ValidateRequiredFlag returns an error if value is empty.
// Use for required command flags.
//
// Example:
//
//	if err := common.ValidateRequiredFlag("--to", toEmail); err != nil {
//	    return err
//	}
func ValidateRequiredFlag(flagName, value string) error {
	if value == "" {
		return NewUserError(
			fmt.Sprintf("%s flag is required", flagName),
			fmt.Sprintf("Use %s <value>", flagName),
		)
	}
	return nil
}

// ValidateRequiredArg returns an error if args is empty or first arg is empty.
// Use for commands that require at least one argument.
//
// Example:
//
//	if err := common.ValidateRequiredArg(args, "message ID"); err != nil {
//	    return err
//	}
func ValidateRequiredArg(args []string, name string) error {
	if len(args) == 0 || args[0] == "" {
		return NewUserError(
			fmt.Sprintf("%s is required", name),
			fmt.Sprintf("Provide %s as an argument", name),
		)
	}
	return nil
}

// ValidateURL returns an error if value is not a valid URL.
//
// Example:
//
//	if err := common.ValidateURL("webhook URL", webhookURL); err != nil {
//	    return err
//	}
func ValidateURL(name, value string) error {
	if value == "" {
		return ValidateRequired(name, value)
	}
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return NewUserError(
			fmt.Sprintf("invalid %s: %s", name, value),
			"URL must be a valid HTTP or HTTPS URL",
		)
	}
	return nil
}

// ValidateEmail returns an error if value doesn't look like an email address.
// This is a basic check for @ symbol, not RFC 5322 compliant.
//
// Example:
//
//	if err := common.ValidateEmail("recipient", toEmail); err != nil {
//	    return err
//	}
func ValidateEmail(name, value string) error {
	if value == "" {
		return ValidateRequired(name, value)
	}
	if !strings.Contains(value, "@") {
		return NewUserError(
			fmt.Sprintf("invalid %s: %s", name, value),
			"Email must contain @ symbol",
		)
	}
	return nil
}

// ValidateOneOf returns an error if value is not in the allowed list.
//
// Example:
//
//	if err := common.ValidateOneOf("status", status, []string{"pending", "active", "cancelled"}); err != nil {
//	    return err
//	}
func ValidateOneOf(name, value string, allowed []string) error {
	if value == "" {
		return nil // Empty is OK, use ValidateRequired for required fields
	}
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return NewUserError(
		fmt.Sprintf("invalid %s: %s", name, value),
		fmt.Sprintf("Allowed values: %s", strings.Join(allowed, ", ")),
	)
}

// ValidateAtLeastOne returns an error if all values are empty.
// Use when at least one of several optional flags is required.
//
// Example:
//
//	if err := common.ValidateAtLeastOne("update field", url, description, status); err != nil {
//	    return err
//	}
func ValidateAtLeastOne(name string, values ...string) error {
	for _, v := range values {
		if v != "" {
			return nil
		}
	}
	return NewUserError(
		fmt.Sprintf("at least one %s is required", name),
		"Provide at least one field to update",
	)
}
