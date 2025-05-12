package config

func IsAuth(id string) *Session {
	session, err := SESSION.GetSession(id)
	if err != nil {
		return nil
	}
	return session
}
