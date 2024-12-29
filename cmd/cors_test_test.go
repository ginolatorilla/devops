package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorsTest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	cmd := newCorsTestCmd()
	cmd.SetArgs([]string{"--target-url", server.URL})

	err := cmd.Execute()

	assert.NoError(t, err)
}

func TestCorsTest_PanicIfResponseCodeIsNon2XX(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	cmd := newCorsTestCmd()
	cmd.SetArgs([]string{"--target-url", server.URL})

	assert.Panics(t, func() {
		cmd.Execute()
	})
}

func TestCorsTest_PanicIfRequiredHeaderIsMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	cmd := newCorsTestCmd()
	cmd.SetArgs([]string{"--target-url", server.URL})

	assert.Panics(t, func() {
		cmd.Execute()
	})
}
