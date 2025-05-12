package utils

import "social/config"

func InitServices() error {
	err := config.ConnectDatabase()
	if err != nil {
		return err
	}
	config.NewSessionManager()
	return nil
}
