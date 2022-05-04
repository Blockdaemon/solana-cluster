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
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
)

type BasicAuth struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

func (b *BasicAuth) Apply(header http.Header) {
	auth := b.Username + ":" + b.Password
	header.Add("authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
}

type BearerAuth struct {
	Token string `json:"token" yaml:"token"`
}

func (b *BearerAuth) Apply(header http.Header) {
	header.Set("authorization", "Bearer "+b.Token)
}

type TLSConfig struct {
	CAFile             string `json:"ca_file" yaml:"ca_file"`
	CertFile           string `json:"cert_file" yaml:"cert_file"`
	KeyFile            string `json:"key_file" yaml:"key_file"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
}

func init() {
	gob.Register(&TLSConfig{})
}

func (t *TLSConfig) Build() (*tls.Config, error) {
	config := &tls.Config{
		InsecureSkipVerify: t.InsecureSkipVerify,
	}
	if len(t.CAFile) > 0 {
		caBytes, err := os.ReadFile(t.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA file: %w", err)
		}
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caBytes) {
			return nil, fmt.Errorf("unable to load CA cert")
		}
		config.RootCAs = caPool
	}
	if len(t.CertFile) > 0 && len(t.KeyFile) == 0 {
		return nil, fmt.Errorf("TLS cert file given but key file missing")
	} else if len(t.CertFile) == 0 && len(t.KeyFile) > 0 {
		return nil, fmt.Errorf("TLS key file given but cert file missing")
	} else if len(t.CertFile) > 0 && len(t.KeyFile) > 0 {
		_, err := t.getClientCertificate(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert and key: %w", err)
		}
		config.GetClientCertificate = t.getClientCertificate
	}
	return config, nil
}

func (t *TLSConfig) getClientCertificate(_ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(t.CertFile, t.KeyFile)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}
