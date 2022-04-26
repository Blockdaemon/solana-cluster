package types

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
)

type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (b *BasicAuth) Apply(req *http.Request) {
	req.SetBasicAuth(b.Username, b.Password)
}

type BearerAuth struct {
	Token string `json:"token"`
}

func (b *BearerAuth) Apply(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+b.Token)
}

type TLSConfig struct {
	CAFile             string `json:"ca_file"`
	CertFile           string `json:"cert_file"`
	KeyFile            string `json:"key_file"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
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
