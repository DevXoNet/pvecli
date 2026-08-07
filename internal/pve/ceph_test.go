// Copyright 2025 DevXo part of vByte Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package pve

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCephAPIEndpoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/nodes/node1/ceph/status":
			_, _ = io.WriteString(w, `{"data":{"health":{"status":"HEALTH_OK"}}}`)
		case "/nodes/node1/ceph/osd":
			_, _ = io.WriteString(w, `{"data":{"nodes":[{"id":0,"status":"up"}]}}`)
		case "/nodes/node1/ceph/pool":
			_, _ = io.WriteString(w, `{"data":[{"pool_name":"ceph-prod"}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client := testClient(server.URL)

	status, err := client.GetCephStatus("node1")
	if err != nil || status["health"] == nil {
		t.Fatalf("GetCephStatus() status = %v, error = %v", status, err)
	}
	osds, err := client.GetCephOSDs("node1")
	if err != nil || osds["nodes"] == nil {
		t.Fatalf("GetCephOSDs() osds = %v, error = %v", osds, err)
	}
	pools, err := client.GetCephPools("node1")
	if err != nil || len(pools) != 1 {
		t.Fatalf("GetCephPools() pools = %v, error = %v", pools, err)
	}
}
