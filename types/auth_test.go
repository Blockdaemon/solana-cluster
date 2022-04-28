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

package types

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicAuth_Apply(t *testing.T) {
	ba := BasicAuth{
		Username: "Aladdin",
		Password: "open sesame",
	}
	header := http.Header{}
	ba.Apply(header)
	assert.Equal(t, "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==", header.Get("Authorization"))
}

func TestBearerAuth_Apply(t *testing.T) {
	ba := BearerAuth{Token: "123"}
	header := http.Header{}
	ba.Apply(header)
	assert.Equal(t, "Bearer 123", header.Get("Authorization"))
}
