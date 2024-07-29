package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/dgrijalva/jwt-go"
)

func setupRouter() *gin.Engine {
	server := NewEventServer()
	return server.SetupRouter()
}

// TestHandleImpression tests the handleImpression handler
func TestHandleImpression(t *testing.T) {
	router := setupRouter()

	// Test case: Missing required parameters
	req, _ := http.NewRequest("GET", "/impression/some-invalid-signed-impression-event", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test case: Valid request
	var impressionEvent = Event {
		UserID:			"a-rondom-token",
		AdID:			"5",
		AdURL:			"google.com",
		PublisherID:	"4",
		EventType:		"impression",
	}
	signedImpressionLink, err := signEvent(&impressionEvent)
	assert.Nil(t, err)
	req, _ = http.NewRequest("GET", "/impression/" + signedImpressionLink, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Impression processed")
}

// TestHandleClick tests the handleClick handler
func TestHandleClick(t *testing.T) {
	router := setupRouter()

	// Test case: Missing required parameters
	req, _ := http.NewRequest("GET", "/click/some-invalid-signed-click-event", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test case: Valid request
	var clickEvent = Event {
		UserID:			"another-rondom-token",
		AdID:			"5",
		AdURL:			"http://yahoo.com",
		PublisherID:	"7",
		EventType:		"click",
	}
	signedClickLink, err := signEvent(&clickEvent)
	assert.Nil(t, err)
	req, _ = http.NewRequest("GET", "/click/" + signedClickLink, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "http://yahoo.com", w.Header().Get("Location"))
}



/* Helper Functions for Testing */

/* Signs the information Event Server needed
 with AdServer's internal private key, so that
 it will be shown that it is really generated by AdServer. */
 func signEvent(event *Event) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, event)
	signedTokenString, err := token.SignedString(JWT_ENCRYPTION_KEY)
	if err != nil {
		return "", err
	}
	return signedTokenString, nil
}