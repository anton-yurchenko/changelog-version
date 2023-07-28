package api

import (
	"changelog-version/utils"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

type API struct {
	Client   Client
	Headers  map[string]string
	Endpoint string
}

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

func New(token string) *API {
	return &API{
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		Headers: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		},
		Endpoint: "https://api.github.com",
	}
}

func (a *API) Do(request *http.Request) (io.ReadCloser, error) {
	for k, v := range a.Headers {
		request.Header.Add(k, v)
	}

	var errr error
	maxRetries := 3
	for i := 1; i <= maxRetries; i++ {
		resp, err := a.Client.Do(request)

		if err != nil || resp.StatusCode != http.StatusOK {
			if i == maxRetries {
				if err != nil {
					errr = err
				} else {
					errr = fmt.Errorf("too many retries")
				}

				break
			}

			delay := math.Pow(3, float64(i+1))
			time.Sleep(time.Duration(delay) * time.Second)
			continue
		}

		return resp.Body, nil
	}

	return nil, errr
}

func (a *API) GetUserEmail(name string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/users/%s", a.Endpoint, name), nil)
	if err != nil {
		return "", utils.Wrap("error creating a api request to fetch author information: %s", err)
	}

	body, err := a.Do(req)
	if err != nil {
		return "", utils.Wrap("error contacting github api: %s", err)
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		return "", utils.Wrap("error reading response body: %s", err)
	}

	s := struct {
		Email string `json:"email"`
	}{}

	if err := json.Unmarshal(data, &s); err != nil {
		return "", utils.Wrap("error decoding response body: %s", err)
	}

	return s.Email, nil
}
