package main

import (
	"fmt"
	"net/http"
	"strings"
)

func (s *App) addAuthKey(key string) error {
	_, err := s.db.Exec("INSERT INTO auth_keys(key) VALUES(?)", key)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return fmt.Errorf("auth key already exists")
		}

		return fmt.Errorf("insert auth key: %w", err)
	}

	fmt.Printf("added auth key: %s\n", key)
	return nil
}

func (s *App) isAuthorized(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return false
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	var exists bool
	err := s.db.QueryRow("SELECT 1 FROM auth_keys WHERE key = ?", token).Scan(&exists)

	return err == nil
}
