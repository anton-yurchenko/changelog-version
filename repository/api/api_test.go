package api_test

import (
	"changelog-version/mocks"
	"changelog-version/repository/api"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	a := assert.New(t)

	t.Log("Test Case 1/1 - OK")
	token := "xxx"

	expected := &api.API{
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		Headers: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		},
		Endpoint: "https://api.github.com",
	}
	got := api.New(token)

	a.Equal(expected, got)
}

func TestDo(t *testing.T) {
	a := assert.New(t)

	type expected struct {
		Result io.ReadCloser
		Error  string
	}

	type mockResult struct {
		output *http.Response
		error  error
	}

	type test struct {
		receiver    *api.API
		mockResults []mockResult
		expected    expected
	}

	suite := map[string]test{
		"OK": {
			receiver: &api.API{
				Headers:  make(map[string]string),
				Endpoint: "http://localhost",
			},
			mockResults: []mockResult{
				{
					output: &http.Response{
						StatusCode: http.StatusOK,
					},
					error: nil,
				},
			},
			expected: expected{
				Result: nil,
				Error:  "",
			},
		},
		"Retry": {
			receiver: &api.API{
				Headers:  make(map[string]string),
				Endpoint: "http://localhost",
			},
			mockResults: []mockResult{
				{
					output: &http.Response{
						StatusCode: http.StatusGatewayTimeout,
					},
					error: fmt.Errorf("reason"),
				},
				{
					output: &http.Response{
						StatusCode: http.StatusOK,
					},
					error: nil,
				},
			},
			expected: expected{
				Result: nil,
				Error:  "",
			},
		},
		"Retries exceeded": {
			receiver: &api.API{
				Headers:  make(map[string]string),
				Endpoint: "http://localhost",
			},
			mockResults: []mockResult{
				{
					output: &http.Response{
						StatusCode: http.StatusGatewayTimeout,
					},
					error: nil,
				},
				{
					output: &http.Response{
						StatusCode: http.StatusGatewayTimeout,
					},
					error: nil,
				},
				{
					output: &http.Response{
						StatusCode: http.StatusGatewayTimeout,
					},
					error: nil,
				},
				{
					output: &http.Response{
						StatusCode: http.StatusGatewayTimeout,
					},
					error: nil,
				},
			},
			expected: expected{
				Result: nil,
				Error:  "too many retries",
			},
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		req, err := http.NewRequest(http.MethodGet, test.receiver.Endpoint, nil)
		if err != nil {
			t.Fatalf("error preparing test: %s", err)
		}

		m := new(mocks.Client)
		test.receiver.Client = m
		for _, mr := range test.mockResults {
			m.On("Do", req).Return(mr.output, mr.error).Once()
		}

		r, err := test.receiver.Do(req)

		a.Equal(test.expected.Result, r)
		if test.expected.Error != "" {
			a.EqualError(err, test.expected.Error)
		} else {
			a.Nil(err)
		}
	}
}

func TestGetUserEmail(t *testing.T) {
	a := assert.New(t)

	type expected struct {
		Result string
		Error  string
	}

	type mockResult struct {
		output *http.Response
		error  error
	}

	type test struct {
		receiver    *api.API
		mockResults mockResult
		expected    expected
	}

	suite := map[string]test{
		"OK": {
			receiver: &api.API{
				Headers:  make(map[string]string),
				Endpoint: "http://localhost",
			},
			mockResults: mockResult{
				output: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"email": "user@domain.com"}`)),
				},
				error: nil,
			},
			expected: expected{
				Result: "user@domain.com",
				Error:  "",
			},
		},
		"Bad response": {
			receiver: &api.API{
				Headers:  make(map[string]string),
				Endpoint: "http://localhost",
			},
			mockResults: mockResult{
				output: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"email": "user@`)),
				},
				error: nil,
			},
			expected: expected{
				Result: "",
				Error:  "error decoding response body: unexpected end of JSON input",
			},
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		m := new(mocks.Client)
		test.receiver.Client = m
		m.On("Do", mock.AnythingOfType("*http.Request")).Return(test.mockResults.output, test.mockResults.error).Once()

		r, err := test.receiver.GetUserEmail("name")

		a.Equal(test.expected.Result, r)
		if test.expected.Error != "" {
			a.EqualError(err, test.expected.Error)
		} else {
			a.Nil(err)
		}
	}
}
