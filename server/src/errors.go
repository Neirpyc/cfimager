package main

import "errors"

const (
//ifPersist = "Please try again, if the error persists, please contact the site owner."
//unexpectedError = "An unexpected error has occurred." + " " + ifPersist
)

func compilationError() error {
	return errors.New("ERR_COMPILATION")
}
