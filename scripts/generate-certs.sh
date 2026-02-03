#!/bin/bash
# Generate self-signed TLS certificates for development

set -e

CERT_DIR="certs"
mkdir -p "$CERT_DIR"

echo "Generating self-signed TLS certificates..."

# Generate private key
openssl genrsa -out "$CERT_DIR/server-key.pem" 2048

# Generate certificate signing request
openssl req -new -key "$CERT_DIR/server-key.pem" -out "$CERT_DIR/server.csr" \
  -subj "/C=US/ST=State/L=City/O=Organization/OU=Unit/CN=localhost"

# Generate self-signed certificate (valid for 365 days)
openssl x509 -req -days 365 -in "$CERT_DIR/server.csr" \
  -signkey "$CERT_DIR/server-key.pem" -out "$CERT_DIR/server-cert.pem"

# Clean up CSR
rm "$CERT_DIR/server.csr"

echo "✓ Certificates generated in $CERT_DIR/"
echo "  - server-cert.pem (certificate)"
echo "  - server-key.pem (private key)"
echo ""
echo "To use TLS, update config/base.yaml:"
echo "  server:"
echo "    tls_cert: \"certs/server-cert.pem\""
echo "    tls_key: \"certs/server-key.pem\""
