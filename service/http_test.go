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

	"github.com/opsee/basic/tp"
	"github.com/opsee/cats/store"
)

func testingCreateAuthToken() string {
	token := base64.StdEncoding.EncodeToString([]byte(`{"id": 1, "customer_id": "e178efbe-8a14-4901-93c0-a455f0b9bce2", "email": "cliff@leaninto.it", "verified": true, "active": true}`))
	return fmt.Sprintf("Basic %s", token)
}

var testingCommon = struct {
	listenAddr string
	svc        *service
	rtr        *tp.Router
}{
	"http://localhost:8080",
	nil,
	nil,
}

func init() {
	svc, _ := NewService(os.Getenv("CATS_POSTGRES_CONN"))
	testingCommon.svc = svc
	if svc != nil {
		testingCommon.rtr = svc.router()
	}
}

func TestGetAssertion(t *testing.T) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/assertions/1", testingCommon.listenAddr), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", testingCreateAuthToken())

	rw := httptest.NewRecorder()

	testingCommon.rtr.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)

	var resp GetChecksResponse

	err = json.Unmarshal(rw.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(resp.Items))
	for _, ca := range resp.Items {
		for _, ass := range ca.Assertions {
			assert.NotNil(t, ass.Key)
		}
	}
}

func TestGetAllAssertions(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/assertions", testingCommon.listenAddr), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", testingCreateAuthToken())

	rw := httptest.NewRecorder()

	testingCommon.rtr.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)

	var resp GetChecksResponse
	err = json.Unmarshal(rw.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(resp.Items))

	for _, ca := range resp.Items {
		for _, ass := range ca.Assertions {
			assert.NotNil(t, ass.Key)
		}
	}
}

func TestPostDeleteAssertions(t *testing.T) {
	assertion := &store.Assertion{
		Key:          "foo",
		Relationship: "notEqual",
		Value:        "",
		Operand:      "quux",
	}

	ca := &CheckAssertions{
		CheckID: "3",
		Assertions: []*store.Assertion{
			assertion,
		},
	}

	caBytes, err := json.Marshal(ca)
	if err != nil {
		t.Fatal(err)
	}
	rdr := bytes.NewBufferString(string(caBytes))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/assertions", testingCommon.listenAddr), rdr)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", testingCreateAuthToken())

	rw := httptest.NewRecorder()
	testingCommon.rtr.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)

	req, err = http.NewRequest("DELETE", fmt.Sprintf("%s/assertions/3", testingCommon.listenAddr), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", testingCreateAuthToken())

	rw = httptest.NewRecorder()

	testingCommon.rtr.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)
}

func TestPutDeleteAssertions(t *testing.T) {
	assertion := &store.Assertion{
		Key:          "foo",
		Relationship: "notEqual",
		Value:        "",
		Operand:      "quux",
	}

	ca := &CheckAssertions{
		CheckID: "3",
		Assertions: []*store.Assertion{
			assertion,
		},
	}

	caBytes, err := json.Marshal(ca)
	if err != nil {
		t.Fatal(err)
	}
	rdr := bytes.NewBufferString(string(caBytes))

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/assertions/3", testingCommon.listenAddr), rdr)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", testingCreateAuthToken())

	rw := httptest.NewRecorder()
	testingCommon.rtr.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)

	req, err = http.NewRequest("DELETE", fmt.Sprintf("%s/assertions/3", testingCommon.listenAddr), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", testingCreateAuthToken())

	rw = httptest.NewRecorder()

	testingCommon.rtr.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)
}
