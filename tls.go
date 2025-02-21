package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"

	"golang.org/x/crypto/pkcs12"
)

func loadTLSCertificate(certPath, password string) (tls.Certificate, error) {
	var cert tls.Certificate
	var err error
	if certPath[:len(certPath)-1] == "fx" || certPath[:len(certPath)-1] == "12" {
		cert, err = loadPKCS12Certificate(certPath, password)
	} else {
		cert, err = tls.LoadX509KeyPair(certPath, certPath)
	}
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("certificate error: %v", err)
	}
	return cert, nil
}

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
