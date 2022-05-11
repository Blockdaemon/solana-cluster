// Copyright 2022 Blockdaemon Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sidecar

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func newRouter(h *SnapshotHandler) http.Handler {
	router := gin.Default()
	h.RegisterHandlers(router)
	return router
}

func testRequest(h *SnapshotHandler, req *http.Request) *httptest.ResponseRecorder {
	router := newRouter(h)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestHandler_ListSnapshots_Error(t *testing.T) {
	h := NewSnapshotHandler("???/some/nonexistent/path", zaptest.NewLogger(t))

	req, err := http.NewRequest(http.MethodGet, "/snapshots", nil)
	require.NoError(t, err)

	res := testRequest(h, req)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}
