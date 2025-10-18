package handlers

import (
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	sessionKeyEntityID      = "SESSION_KEY_ENTITY_ID"
	sessionKeyEngineVersion = "SESSION_KEY_ENGINE_VERSION"
	sessionKeyPatchVersion  = "SESSION_KEY_PATCH_VERSION"
	sessionKeyUserID        = "SESSION_KEY_USER_ID"
	sessionKeyIsPC          = "SESSION_KEY_IS_PC"
)

var sessionStore sync.Map // map[string]*sync.Map

func getSessionByBearer(c *gin.Context) *sync.Map {
	bearer := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if bearer == "" {
		return nil
	}
	if value, ok := sessionStore.Load(bearer); ok {
		return value.(*sync.Map)
	}
	session := &sync.Map{}
	sessionStore.Store(bearer, session)
	return session
}

func resetSession(bearer string) {
	if bearer == "" {
		return
	}
	sessionStore.Delete(bearer)
}
