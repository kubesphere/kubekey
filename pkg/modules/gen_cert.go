package modules

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math"
	"math/big"
	"net"
	"time"

	"github.com/cockroachdb/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	cgutilcert "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"k8s.io/klog/v2"
	netutils "k8s.io/utils/net"
	"k8s.io/utils/ptr"

	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

/*
The GenCert module is designed to generate SSL/TLS certificates for secure communications.
It supports both self-signed certificates and certificates signed by a root Certificate Authority (CA).

Configuration:
You can customize certificate generation with the following parameters:

gen_cert:
  cn: example.com             # required: Common Name for the certificate
  out_key: /path/to/key       # required: Output path for the private key
  out_cert: /path/to/cert     # required: Output path for the certificate
  root_key: /path/to/ca.key   # optional: Path to the root CA private key
  root_cert: /path/to/ca.crt  # optional: Path to the root CA certificate
  sans:                       # optional: Subject Alternative Names (SANs)
    - example.com
    - www.example.com
  policy: IfNotPresent        # optional: Certificate generation policy
  date: 8760h                 # optional: Certificate validity period

Usage Examples in Playbook Tasks:
1. Generate a self-signed certificate:
   ```yaml
   - name: Generate a self-signed certificate
     gen_cert:
       cn: example.com
       out_key: /etc/ssl/private/example.key
       out_cert: /etc/ssl/certs/example.crt
       sans:
         - example.com
         - www.example.com
     register: cert_result
   ```

2. Generate a certificate signed by a root CA:
   ```yaml
   - name: Generate a certificate signed by a root CA
     gen_cert:
       cn: example.com
       root_key: /etc/ssl/private/ca.key
       root_cert: /etc/ssl/certs/ca.crt
       out_key: /etc/ssl/private/example.key
       out_cert: /etc/ssl/certs/example.crt
     register: signed_cert
   ```

Return Values:
- On success: "Success" is returned in stdout.
- On failure: An error message is returned in stderr.
*/

const (
	// defaultSignCertAfter specifies the default validity period for signed certificates (10 years).
	defaultSignCertAfter = time.Hour * 24 * 365 * 10
	// certificateBlockType is the PEM block type for certificates.
	certificateBlockType = "CERTIFICATE"
	rsaKeySize           = 2048

	// Certificate generation policies:
	// policyAlways: Always generate a new certificate, overwriting any existing one.
	policyAlways = "Always"
	// policyIfNotPresent: If a certificate exists, validate it. If validation fails or it doesn't exist, generate a new one.
	policyIfNotPresent = "IfNotPresent"
	// policyNone: Only validate the certificate; do not generate a new one.
	policyNone = "None"
)

// defaultAltName provides default SANs for certificates.
var defaultAltName = &cgutilcert.AltNames{
	DNSNames: []string{"localhost"},
	IPs:      []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
}

// genCertArgs holds the arguments for certificate generation.
type genCertArgs struct {
	rootKey  string
	rootCert string
	date     time.Duration
	policy   string
	sans     []string
	cn       string
	outKey   string
	outCert  string
	isCA     *bool
}

// signedCertificate generates a certificate signed by the specified root CA.
func (gca genCertArgs) signedCertificate(cfg cgutilcert.Config) (string, string, error) {
	// Load the CA private key.
	caKey, err := TryLoadKeyFromDisk(gca.rootKey)
	if err != nil {
		return StdoutFailed, "Failed to load root key", err
	}
	// Load the CA certificate chain.
	caCert, err := TryLoadCertChainFromDisk(gca.rootCert)
	if err != nil {
		return StdoutFailed, "Failed to load root certificate", err
	}

	// Helper function to generate and write a new certificate and key.
	generateAndWrite := func() (string, string, error) {
		newKey, err := rsa.GenerateKey(cryptorand.Reader, rsaKeySize)
		if err != nil {
			return StdoutFailed, "Failed to generate RSA key", err
		}
		newCert, err := NewSignedCert(cfg, gca.date, newKey, caCert[0], caKey, ptr.Deref(gca.isCA, false))
		if err != nil {
			return StdoutFailed, "Failed to generate signed certificate", err
		}
		if err := WriteKey(gca.outKey, newKey, gca.policy); err != nil {
			return StdoutFailed, "Failed to write private key", err
		}
		if err := WriteCert(gca.outCert, newCert, gca.policy); err != nil {
			return StdoutFailed, "Failed to write certificate", err
		}
		return StdoutSuccess, "", nil
	}

	// Helper function to verify the existing certificate and key.
	verify := func() error {
		// Check if the private key exists and is valid.
		if _, err := TryLoadKeyFromDisk(gca.outKey); err != nil {
			return err
		}
		// Check if the certificate exists and is valid.
		existCert, err := TryLoadCertChainFromDisk(gca.outCert)
		if err != nil {
			return err
		}
		// Validate the certificate's validity period.
		if err := ValidateCertPeriod(existCert[0], 0); err != nil {
			return err
		}
		// Validate the certificate chain.
		if err := VerifyCertChain(existCert[0], existCert[:1], caCert[0]); err != nil {
			return err
		}
		// Validate the certificate's SANs and other configuration.
		return validateCertificateWithConfig(existCert[0], gca.outCert, cfg)
	}

	switch gca.policy {
	case policyAlways:
		// For all other cases (including policyAlways), always generate a new certificate and key.
		return generateAndWrite()
	case policyIfNotPresent:
		if err := verify(); err != nil {
			klog.V(4).ErrorS(err, "Certificate or key verification failed, will regenerate", "outKey", gca.outKey, "outCert", gca.outCert)
			return generateAndWrite()
		}
		// Existing certificate and key are valid; skip generation.
		return StdoutSkip, "", nil
	case policyNone:
		if err := verify(); err != nil {
			return StdoutFailed, "Certificate validation failed", err
		}
		return StdoutSkip, "", nil
	default:
		return StdoutFailed, "unsupport policy", errors.New("unsupport policy")
	}
}

// selfSignedCertificate creates a self-signed certificate and writes it to disk according to the specified policy.
// It returns a status string, an optional message, and an error if one occurred.
func (gca genCertArgs) selfSignedCertificate(cfg cgutilcert.Config) (string, string, error) {
	// Generates a new self-signed certificate and writes both the key and certificate to their respective files.
	generateAndWrite := func() (string, string, error) {
		newKey, err := rsa.GenerateKey(cryptorand.Reader, rsaKeySize)
		if err != nil {
			return StdoutFailed, "Unable to generate RSA private key", err
		}

		newCert, err := NewSelfSignedCACert(cfg, gca.date, newKey)
		if err != nil {
			return StdoutFailed, "Unable to generate self-signed certificate", err
		}

		// Persist the private key and certificate to disk.
		if err := WriteKey(gca.outKey, newKey, gca.policy); err != nil {
			return StdoutFailed, "Unable to write private key to file", err
		}
		if err := WriteCert(gca.outCert, newCert, gca.policy); err != nil {
			return StdoutFailed, "Unable to write certificate to file", err
		}

		return StdoutSuccess, "", nil
	}

	// Verifies that both the private key and certificate exist and are valid.
	verify := func() error {
		if _, err := TryLoadKeyFromDisk(gca.outKey); err != nil {
			return err
		}
		if _, err := TryLoadCertChainFromDisk(gca.outCert); err != nil {
			return err
		}
		return nil
	}

	switch gca.policy {
	case policyAlways:
		// Always generate a new certificate and key, regardless of existing files.
		return generateAndWrite()
	case policyIfNotPresent:
		// If verification fails, log and regenerate; otherwise, skip generation.
		if err := verify(); err != nil {
			klog.V(4).ErrorS(err, "Existing self-signed certificate or key is invalid or missing, regenerating", "outKey", gca.outKey, "outCert", gca.outCert)
			return generateAndWrite()
		}
		return StdoutSkip, "", nil
	case policyNone:
		// Only verify the presence and validity of the certificate and key.
		if err := verify(); err != nil {
			return StdoutFailed, "Self-signed certificate or key validation failed", err
		}
		return StdoutSkip, "", nil
	default:
		return StdoutFailed, "unsupported policy", errors.New("unsupported policy")
	}
}

// newGenCertArgs parses and validates the arguments for certificate generation.
func newGenCertArgs(_ context.Context, raw runtime.RawExtension, vars map[string]any) (*genCertArgs, error) {
	gca := &genCertArgs{}
	// Parse arguments.
	args := variable.Extension2Variables(raw)
	gca.rootKey, _ = variable.StringVar(vars, args, "root_key")
	gca.rootCert, _ = variable.StringVar(vars, args, "root_cert")
	gca.date, _ = variable.DurationVar(vars, args, "date")
	gca.policy, _ = variable.StringVar(vars, args, "policy")
	gca.sans, _ = variable.StringSliceVar(vars, args, "sans")
	gca.cn, _ = variable.StringVar(vars, args, "cn")
	gca.outKey, _ = variable.StringVar(vars, args, "out_key")
	gca.outCert, _ = variable.StringVar(vars, args, "out_cert")
	gca.isCA, _ = variable.BoolVar(vars, args, "is_ca")
	// Validate arguments.
	if gca.policy != policyAlways && gca.policy != policyIfNotPresent && gca.policy != policyNone {
		return nil, errors.New("\"policy\" must be one of [Always, IfNotPresent, None]")
	}
	if gca.outKey == "" || gca.outCert == "" {
		return nil, errors.New("\"out_key\" and \"out_cert\" must be specified as strings")
	}
	if gca.cn == "" {
		return nil, errors.New("\"cn\" must be specified as a string")
	}

	return gca, nil
}

// ModuleGenCert is the entry point for the "gen_cert" module, responsible for generating SSL/TLS certificates.
func ModuleGenCert(ctx context.Context, options ExecOptions) (string, string, error) {
	// Retrieve all host variables.
	ha, err := options.getAllVariables()
	if err != nil {
		return StdoutFailed, StderrGetHostVariable, err
	}

	gca, err := newGenCertArgs(ctx, options.Args, ha)
	if err != nil {
		return StdoutFailed, StderrParseArgument, err
	}

	cfg := &cgutilcert.Config{
		CommonName:   gca.cn,
		Organization: []string{"kubekey"},
		AltNames:     appendSANsToAltNames(defaultAltName, gca.sans),
	}

	switch {
	case gca.rootKey == "" || gca.rootCert == "":
		return gca.selfSignedCertificate(*cfg)
	default:
		return gca.signedCertificate(*cfg)
	}
}

// WriteKey writes the given private key to the specified file path.
func WriteKey(outKey string, key crypto.Signer, policy string) error {
	if key == nil {
		return errors.New("private key cannot be nil when writing to file")
	}

	encoded, err := keyutil.MarshalPrivateKeyToPEM(key)
	if err != nil {
		return errors.Wrap(err, "failed to marshal private key to PEM")
	}
	if err := keyutil.WriteKey(outKey, encoded); err != nil {
		return errors.Wrapf(err, "failed to write private key to file %s", outKey)
	}

	return nil
}

// WriteCert writes the given certificate to the specified file path.
func WriteCert(outCert string, cert *x509.Certificate, policy string) error {
	if cert == nil {
		return errors.New("certificate cannot be nil when writing to file")
	}

	if err := cgutilcert.WriteCert(outCert, EncodeCertPEM(cert)); err != nil {
		return errors.Wrapf(err, "failed to write certificate to file %s", outCert)
	}

	return nil
}

// EncodeCertPEM encodes the given certificate into PEM format.
func EncodeCertPEM(cert *x509.Certificate) []byte {
	block := pem.Block{
		Type:  certificateBlockType,
		Bytes: cert.Raw,
	}

	return pem.EncodeToMemory(&block)
}

// TryLoadKeyFromDisk attempts to load and validate a private key from disk.
func TryLoadKeyFromDisk(rootKey string) (crypto.Signer, error) {
	// Parse the private key from the specified file.
	privKey, err := keyutil.PrivateKeyFromFile(rootKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load the private key file %s", rootKey)
	}

	// Only RSA and ECDSA private keys are supported.
	var key crypto.Signer
	switch k := privKey.(type) {
	case *rsa.PrivateKey:
		key = k
	case *ecdsa.PrivateKey:
		key = k
	default:
		return nil, errors.Errorf("the private key file %s is neither in RSA nor ECDSA format", rootKey)
	}

	return key, nil
}

// TryLoadCertChainFromDisk loads a certificate chain from the specified file.
func TryLoadCertChainFromDisk(rootCert string) ([]*x509.Certificate, error) {
	return cgutilcert.CertsFromFile(rootCert)
}

// appendSANsToAltNames parses SANs from a list of strings and adds them to altNames for use in a certificate.
// Valid IP addresses are added to altNames.IPs, and valid DNS names (including wildcards) are added to altNames.DNSNames.
// Invalid entries are logged as warnings.
func appendSANsToAltNames(altNames *cgutilcert.AltNames, sans []string) cgutilcert.AltNames {
	for _, altname := range sans {
		if ip := netutils.ParseIPSloppy(altname); ip != nil {
			altNames.IPs = append(altNames.IPs, ip)
		} else if len(validation.IsDNS1123Subdomain(altname)) == 0 {
			altNames.DNSNames = append(altNames.DNSNames, altname)
		} else if len(validation.IsWildcardDNS1123Subdomain(altname)) == 0 {
			altNames.DNSNames = append(altNames.DNSNames, altname)
		} else {
			klog.V(4).Infof(
				"[certificates] WARNING: Failed to add '%s' to the SAN list, as it is not a valid IP or RFC-1123-compliant DNS entry\n",
				altname,
			)
		}
	}

	return *altNames
}

// NewSelfSignedCACert creates a new self-signed CA certificate.
func NewSelfSignedCACert(cfg cgutilcert.Config, after time.Duration, key crypto.Signer) (*x509.Certificate, error) {
	now := time.Now()
	// Generate a random serial number in the range [1, MaxInt64).
	serial, err := cryptorand.Int(cryptorand.Reader, new(big.Int).SetInt64(math.MaxInt64-1))
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate certificate serial number")
	}
	serial = new(big.Int).Add(serial, big.NewInt(1))

	notBefore := now.UTC()
	if !cfg.NotBefore.IsZero() {
		notBefore = cfg.NotBefore.UTC()
	}
	if after == 0 { // Default validity: 10 years.
		after = defaultSignCertAfter
	}

	tmpl := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:              []string{cfg.CommonName},
		NotBefore:             notBefore,
		NotAfter:              now.Add(after).UTC(),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &tmpl, &tmpl, key.Public(), key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create certificate")
	}

	return x509.ParseCertificate(certDERBytes)
}

// NewSignedCert creates a certificate signed by the given CA certificate and key.
func NewSignedCert(cfg cgutilcert.Config, after time.Duration, key crypto.Signer, caCert *x509.Certificate, caKey crypto.Signer, isCA bool) (*x509.Certificate, error) {
	// Generate a random serial number in the range [1, MaxInt64).
	serial, err := cryptorand.Int(cryptorand.Reader, new(big.Int).SetInt64(math.MaxInt64-1))
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate certificate serial number")
	}
	serial = new(big.Int).Add(serial, big.NewInt(1))

	if cfg.CommonName == "" {
		return nil, errors.New("commonName must be specified")
	}

	keyUsage := x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	if isCA {
		keyUsage |= x509.KeyUsageCertSign
	}

	RemoveDuplicateAltNames(&cfg.AltNames)

	if after == 0 {
		after = defaultSignCertAfter
	}

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:              cfg.AltNames.DNSNames,
		IPAddresses:           cfg.AltNames.IPs,
		SerialNumber:          serial,
		NotBefore:             caCert.NotBefore,
		NotAfter:              time.Now().Add(after).UTC(),
		KeyUsage:              keyUsage,
		ExtKeyUsage:           cfg.Usages,
		BasicConstraintsValid: true,
		IsCA:                  isCA,
	}

	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create certificate")
	}

	return x509.ParseCertificate(certDERBytes)
}

// RemoveDuplicateAltNames eliminates duplicate entries from the AltNames struct.
func RemoveDuplicateAltNames(altNames *cgutilcert.AltNames) {
	if altNames == nil {
		return
	}

	if altNames.DNSNames != nil {
		altNames.DNSNames = sets.List(sets.New(altNames.DNSNames...))
	}

	ipsKeys := make(map[string]struct{})
	var ips []net.IP
	for _, one := range altNames.IPs {
		if _, ok := ipsKeys[one.String()]; !ok {
			ipsKeys[one.String()] = struct{}{}
			ips = append(ips, one)
		}
	}
	altNames.IPs = ips
}

// ValidateCertPeriod checks whether the certificate is currently valid, considering the given offset.
func ValidateCertPeriod(cert *x509.Certificate, offset time.Duration) error {
	period := fmt.Sprintf("NotBefore: %v, NotAfter: %v", cert.NotBefore, cert.NotAfter)
	now := time.Now().Add(offset)
	if now.Before(cert.NotBefore) {
		return errors.Errorf("the certificate is not valid yet: %s", period)
	}
	if now.After(cert.NotAfter) {
		return errors.Errorf("the certificate has expired: %s", period)
	}

	return nil
}

// VerifyCertChain ensures that a certificate has a valid chain of trust back to the root CA.
func VerifyCertChain(cert *x509.Certificate, intermediates []*x509.Certificate, root *x509.Certificate) error {
	rootPool := x509.NewCertPool()
	rootPool.AddCert(root)

	intermediatePool := x509.NewCertPool()
	for _, c := range intermediates {
		intermediatePool.AddCert(c)
	}

	verifyOptions := x509.VerifyOptions{
		Roots:         rootPool,
		Intermediates: intermediatePool,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	if _, err := cert.Verify(verifyOptions); err != nil {
		return errors.Wrapf(err, "failed to verify certificate")
	}

	return nil
}

// validateCertificateWithConfig ensures that the certificate is valid for all SANs specified in the configuration.
func validateCertificateWithConfig(cert *x509.Certificate, baseName string, cfg cgutilcert.Config) error {
	for _, dnsName := range cfg.AltNames.DNSNames {
		if err := cert.VerifyHostname(dnsName); err != nil {
			return errors.Wrapf(err, "certificate %s is invalid", baseName)
		}
	}

	for _, ipAddress := range cfg.AltNames.IPs {
		if err := cert.VerifyHostname(ipAddress.String()); err != nil {
			return errors.Wrapf(err, "certificate %s is invalid", baseName)
		}
	}

	return nil
}

func init() {
	utilruntime.Must(RegisterModule("gen_cert", ModuleGenCert))
}
