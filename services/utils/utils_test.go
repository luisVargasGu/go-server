package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendJSONResponse(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		data         interface{}
		expectedBody string
	}{
		{
			name:       "Success response",
			statusCode: http.StatusOK,
			data:       map[string]string{"message": "success"},
			expectedBody: `{"message":"success"}`,
		},
		{
			name:       "Error response",
			statusCode: http.StatusInternalServerError,
			data:       map[string]string{"error": "internal server error"},
			expectedBody: `{"error":"internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the function
			SendJSONResponse(rr, tt.statusCode, tt.data)

			// Check the status code
			if status := rr.Code; status != tt.statusCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.statusCode)
			}

			// Check the Content-Type header
			if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("handler returned wrong content type: got %v want %v",
					contentType, "application/json")
			}

			// Check the response body
			var responseBody map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
			if err != nil {
				t.Fatalf("could not unmarshal response body: %v", err)
			}

			var expectedBody map[string]interface{}
			err = json.Unmarshal([]byte(tt.expectedBody), &expectedBody)
			if err != nil {
				t.Fatalf("could not unmarshal expected body: %v", err)
			}

			if len(responseBody) != len(expectedBody) {
				t.Errorf("response body length mismatch: got %v want %v",
					len(responseBody), len(expectedBody))
			}

			for k, v := range expectedBody {
				if responseBody[k] != v {
					t.Errorf("response body mismatch for key %v: got %v want %v",
						k, responseBody[k], v)
				}
			}
		})
	}
}

func TestParseJson(t *testing.T) {
	type test1 struct {
		Field int `json:"field"`
	}

	type test2 struct {
		Name string `json:"name"`
	}

	tests := []struct {
		body string
		name string
		dtype any
		expected any
	}{
		{
			name: "Success 1",
			body: `{ "field": 10 }`,
			dtype: &test1{},
			expected: test1{ Field: 10 },
		},
		{
			name: "Success 2",
			body: `{ "name": "test2" }`,
			dtype: &test2{},
			expected: test2{ Name: "test2" },
		},
		{
			name:     "Missing Body",
			body:     "",
			dtype:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func ( *testing.T ) {
			var r *http.Request
			if tt.body == "" {
				r = httptest.NewRequest("POST", "/test", nil)
			} else {
				r = httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))
			}

			err := ParseJSON(r, tt.dtype)
			if err != nil && tt.dtype != nil {
				t.Fatalf("Could not parse body to JSON: %v", err)
			}

			if !compare(tt.expected, tt.dtype) {
				t.Fatalf("Parsed JSON is not equal. Was %v Expected %v.", tt.dtype, tt.expected)
			}
		})
	}
}

func compare(a, b any) bool {
	return jsonEqual(a, b)
}

func jsonEqual(a, b any) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return string(aj) == string(bj)
}
