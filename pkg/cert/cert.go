package cert

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"path"
	"time"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	_projectName           = "sidecar-go"
	_webhookObjectMetaName = "sidecar-go-webhook-configuration"
	_webhookName           = "sidecar-go.togettoyou.com"
)

type Manager struct {
	Client            client.Client
	CertDir           string
	WebhookURL        string
	WebhookInjectPath string
	ServiceName       string
	Namespace         string
	orgs              []string
	commonName        string
	dnsNames          []string
}

func Init(m *Manager) error {
	m.orgs = []string{_projectName}
	m.commonName = _projectName
	m.dnsNames = []string{fmt.Sprintf("%s.%s.svc", m.ServiceName, m.Namespace)}

	if m.WebhookURL != "" {
		u, err := url.Parse(m.WebhookURL)
		if err != nil {
			return err
		}
		m.dnsNames = append(m.dnsNames, u.Hostname())
	}

	caPEM, err := m.createCert()
	if err != nil {
		return err
	}

	return m.createMutatingWebhookConfiguration(caPEM)
}

func (m *Manager) createCert() (*bytes.Buffer, error) {
	ca := &x509.Certificate{
		SerialNumber:          big.NewInt(2048),
		Subject:               pkix.Name{Organization: m.orgs},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, err
	}

	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return nil, err
	}

	cert := &x509.Certificate{
		DNSNames:     m.dnsNames,
		SerialNumber: big.NewInt(1024),
		Subject: pkix.Name{
			CommonName:   m.commonName,
			Organization: m.orgs,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	serverPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	serverCertBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &serverPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, err
	}

	serverCertPEM := new(bytes.Buffer)
	err = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})
	if err != nil {
		return nil, err
	}

	serverPrivateKeyPEM := new(bytes.Buffer)
	err = pem.Encode(serverPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey),
	})
	if err != nil {
		return nil, err
	}

	fmt.Println(m.CertDir)
	err = os.MkdirAll(m.CertDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	err = writeFile(path.Join(m.CertDir, "tls.crt"), serverCertPEM)
	if err != nil {
		return nil, err
	}
	err = writeFile(path.Join(m.CertDir, "tls.key"), serverPrivateKeyPEM)
	if err != nil {
		return nil, err
	}

	return caPEM, nil
}

func (m *Manager) createMutatingWebhookConfiguration(caPEM *bytes.Buffer) error {
	clientConfig := admissionregistrationv1.WebhookClientConfig{
		CABundle: caPEM.Bytes(),
	}
	if m.WebhookURL != "" {
		clientConfig.URL = &m.WebhookURL
	} else {
		clientConfig.Service = &admissionregistrationv1.ServiceReference{
			Name:      m.ServiceName,
			Namespace: m.Namespace,
			Path:      &m.WebhookInjectPath,
		}
	}

	mutatingWebhookConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: _webhookObjectMetaName,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{{
			Name:                    _webhookName,
			AdmissionReviewVersions: []string{"v1"},
			SideEffects: func() *admissionregistrationv1.SideEffectClass {
				se := admissionregistrationv1.SideEffectClassNone
				return &se
			}(),
			ClientConfig: clientConfig,
			Rules: []admissionregistrationv1.RuleWithOperations{
				{
					Operations: []admissionregistrationv1.OperationType{
						admissionregistrationv1.Create,
						admissionregistrationv1.Update,
					},
					Rule: admissionregistrationv1.Rule{
						APIGroups:   []string{""},
						APIVersions: []string{"v1"},
						Resources:   []string{"pods"},
					},
				},
			},
			FailurePolicy: func() *admissionregistrationv1.FailurePolicyType {
				pt := admissionregistrationv1.Fail
				return &pt
			}(),
		}},
	}

	_ = m.Client.Delete(context.Background(), mutatingWebhookConfig)
	return m.Client.Create(context.Background(), mutatingWebhookConfig)
}

func writeFile(filepath string, content *bytes.Buffer) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(content.Bytes())
	if err != nil {
		return err
	}
	return nil
}
