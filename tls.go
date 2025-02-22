package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"

	"golang.org/x/crypto/pkcs12"
)

func loadPKCS12Certificate(certPath, password string) (tls.Certificate, error) {
	fileData, err := os.ReadFile(certPath)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("cannot read PKCS#12: %v", err)
	}

	privateKey, certData, err := pkcs12.Decode(fileData, password)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("cannot decode PKCS#12: %v", err)
	}

	certDer := certData.Raw
	tlsCert := tls.Certificate{
		Certificate: [][]byte{certDer},
		PrivateKey:  privateKey,
	}

	return tlsCert, nil
}

func isTLSHandshakeFailure(err error) bool {
	var tlsErr *tls.RecordHeaderError
	if err == nil {
		return false
	}
	if errors.As(err, &tlsErr) {
		return true
	}
	return err.Error() == "EOF"
}
