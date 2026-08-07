// Copyright 2025 DevXo part of vByte Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package pve

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func testClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		TokenID: "user@pam!test",
		Secret:  "secret",
		client:  &http.Client{Timeout: time.Second},
	}
}

func TestDoRequestSetsAuthorizationAndDecodesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Authorization"), "PVEAPIToken=user@pam!test=secret"; got != want {
			t.Errorf("Authorization = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":"ok"}`)
	}))
	defer server.Close()

	var response struct {
		Data string `json:"data"`
	}
	if err := testClient(server.URL).doRequest(http.MethodGet, "/test", &response); err != nil {
		t.Fatalf("doRequest() error = %v", err)
	}
	if response.Data != "ok" {
		t.Fatalf("Data = %q, want ok", response.Data)
	}
}

func TestDoRequestWithDataEncodesFormValues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		values, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatal(err)
		}
		if got, want := values.Get("command"), "echo a&b + 100%"; got != want {
			t.Errorf("command = %q, want %q", got, want)
		}
		_, _ = io.WriteString(w, `{"data":null}`)
	}))
	defer server.Close()

	err := testClient(server.URL).doRequestWithData(http.MethodPost, "/test", map[string]string{
		"command": "echo a&b + 100%",
	}, nil)
	if err != nil {
		t.Fatalf("doRequestWithData() error = %v", err)
	}
}

func TestDoRequestRejectsInvalidURL(t *testing.T) {
	err := testClient("://invalid").doRequest(http.MethodGet, "/test", nil)
	if err == nil || !strings.Contains(err.Error(), "create request") {
		t.Fatalf("doRequest() error = %v, want create request error", err)
	}
}

func TestClientInitializationErrorIsReturned(t *testing.T) {
	want := errors.New("invalid configuration")
	client := &Client{initErr: want}
	if err := client.doRequest(http.MethodGet, "/test", nil); !errors.Is(err, want) {
		t.Fatalf("doRequest() error = %v, want %v", err, want)
	}
	if err := client.doRequestWithData(http.MethodPost, "/test", nil, nil); !errors.Is(err, want) {
		t.Fatalf("doRequestWithData() error = %v, want %v", err, want)
	}
}
