package types_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	fmt "fmt"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"https://github.com/unigrid-project/cosmos-sdk-ugdmint"
)

const (
	json string = "{" +
		"\"data\": {" +
		"\"mints\":[{" +
		"\"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43hlvy4jatsg/80\":\"100\"," +
		"\"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43tlvy4jatsg/90\":\"1000\"," +
		"\"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43ulvy4jatsg/110\":\"1275\"," +
		"\"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43alvy4jatsg/150\":\"981256\"," +
		"\"unigrid1pk2sxhrywmxsqtnas3p7gu0t8x43rlvy4jatsg/165\":\"1236\"," +
		"\"pk2sxhrywmxsqtnas3p7gu0t8x43rlvy4jatsg/147621207\":\"1236\"," +
		"\"pk2sxhrywmrsqtnas3p7gu0t8x43rlvy4jatsg/1238965123\":\"1236\"" +
		"}]" +
		"}" +
		"}"
)

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

func TestCanMintFromHedgehog(t *testing.T) {

	// priv, err := rsa.GenerateKey(rand.Reader, *rsaBits)
	priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	/*
	   hosts := strings.Split(*host, ",")
	   for _, h := range hosts {
	   	if ip := net.ParseIP(h); ip != nil {
	   		template.IPAddresses = append(template.IPAddresses, ip)
	   	} else {
	   		template.DNSNames = append(template.DNSNames, h)
	   	}
	   }
	   if *isCA {
	   	template.IsCA = true
	   	template.KeyUsage |= x509.KeyUsageCertSign
	   }
	*/

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)

	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	out := &bytes.Buffer{}
	out2 := &bytes.Buffer{}
	pem.Encode(out, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certKey := out.Bytes()
	fmt.Println(out.String())
	out.Reset()
	pem.Encode(out2, pemBlockForKey(priv))
	pubKey := out2.Bytes()
	fmt.Println(out2.String())

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(json))
	}))
	server.URL = "https://127.0.0.1:52884/mint-storage"
	cert, err := tls.X509KeyPair(certKey, pubKey)
	//tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Panic("bad server certs: ", err)
	}
	certs := []tls.Certificate{cert}
	server.TLS = &tls.Config{Certificates: certs}

	server.StartTLS()

	defer server.Close()
	cache := types.minter.newCache()

	cache.callHedgehog("mint-storage")
}
