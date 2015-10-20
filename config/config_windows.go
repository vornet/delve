package config

import (
	"syscall"
)

func getHomeDir() (string, error) {
	// TODO: This is a workaround for user.Current being
	// very slow on a domain joined PC that is not connected
	// to the domain.
	t, e := syscall.OpenCurrentProcessToken()
	if e != nil {
		return "", e
	}
	defer t.Close()
	dir, e := t.GetUserProfileDirectory()
	if e != nil {
		return "", e
	}
	return dir, nil
}
