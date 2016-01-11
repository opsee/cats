package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	//"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/opsee/cats/checker"
)

func testingCreateAuthToken() string {
	token := base64.StdEncoding.EncodeToString([]byte(`{"id": 1, "customer_id": "e178efbe-8a14-4901-93c0-a455f0b9bce2", "email": "cliff@leaninto.it", "verified": true, "active": true}`))
	return fmt.Sprintf("Basic %s", token)
}

func TestGetChecks(t *testing.T) {
	listenAddr := "http://localhost:8080"
	svc, err := NewService(os.Getenv("POSTGRES_CONN"))
	if err != nil {
		t.Fatal(err)
	}

	rtr := svc.router()

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/assertions?checks=1,2", listenAddr), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", testingCreateAuthToken())

	rw := httptest.NewRecorder()

	rtr.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)

	assertions := []*checker.Assertion{}

	err = json.Unmarshal(rw.Body.Bytes(), &assertions)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(assertions))
	for _, ass := range assertions {
		assert.NotNil(t, ass.Key)
	}
}

func TestPutDeleteAssertions(t *testing.T) {
	listenAddr := "http://localhost:8080"
	svc, err := NewService(os.Getenv("POSTGRES_CONN"))
	if err != nil {
		t.Fatal(err)
	}

	rtr := svc.router()
	assertion := &checker.Assertion{
		Key:          "foo",
		Relationship: "notEqual",
		Value:        "",
		Operand:      "quux",
	}

	check := &checker.Check{
		Id: "3",
		Assertions: []*checker.Assertion{
			assertion,
		},
	}

	checkRequest := &PutCheckRequest{
		Check: check,
	}

	reqBytes, err := json.Marshal(checkRequest)
	if err != nil {
		t.Fatal(err)
	}

	rdr := bytes.NewBufferString(string(reqBytes))
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/assertions", listenAddr), rdr)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", testingCreateAuthToken())

	rw := httptest.NewRecorder()

	rtr.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)

	rdr = bytes.NewBufferString(string(reqBytes))
	req, err = http.NewRequest("DELETE", fmt.Sprintf("%s/assertions", listenAddr), rdr)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", testingCreateAuthToken())

	rw = httptest.NewRecorder()

	rtr.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)
}
