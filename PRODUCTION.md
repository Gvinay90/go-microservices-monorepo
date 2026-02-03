# Production Deployment Guide

This guide covers deploying the expense-sharing application to production with enterprise-grade security.

## Security Features Implemented

✅ **Authentication**: Keycloak JWT validation  
✅ **Authorization**: Role-based access control (RBAC)  
✅ **TLS/HTTPS**: Encrypted communication  
✅ **Rate Limiting**: Token bucket algorithm (100 req/min per user)  
✅ **Audit Logging**: Security event tracking  
✅ **Graceful Shutdown**: Zero-downtime deployments  
✅ **Health Checks**: Kubernetes-ready endpoints  

---

## 1. TLS/HTTPS Configuration

### Generate Certificates

**For Development:**
```bash
./scripts/generate-certs.sh
```

**For Production:**
Use certificates from a trusted CA (Let's Encrypt, DigiCert, etc.)

### Enable TLS

Edit `config/base.yaml`:
```yaml
server:
  tls_cert: "certs/server-cert.pem"
  tls_key: "certs/server-key.pem"
```

### Test TLS Connection

```bash
# With grpcurl (requires -insecure for self-signed certs)
grpcurl -insecure \
  -H "authorization: Bearer $TOKEN" \
  localhost:50051 user.UserService/ListUsers
```

---

## 2. Rate Limiting

Protects against abuse and DDoS attacks using token bucket algorithm.

### Configuration

```yaml
server:
  rate_limit: 100  # requests per minute per user
```

### How It Works

- Each user gets 100 tokens per minute
- Each request consumes 1 token
- Tokens refill every minute
- Anonymous users are rate-limited by IP address

### Response When Limited

```
Code: ResourceExhausted
Message: "rate limit exceeded, please try again later"
```

---

## 3. Audit Logging

Logs all security-relevant events for compliance and forensics.

### Enable Audit Logging

```yaml
auth:
  audit_logging: true
```

### What Gets Logged

- ✅ Authentication attempts (success/failure)
- ✅ Authorization failures
- ✅ User ID, email, roles
- ✅ Client IP address
- ✅ Request method and duration
- ✅ Timestamps

### Example Log Entry

```json
{
  "time": "2026-01-08T11:30:00Z",
  "level": "INFO",
  "msg": "Authenticated request",
  "event": "grpc_request",
  "method": "/user.UserService/RegisterUser",
  "user_id": "abc-123",
  "email": "alice@example.com",
  "roles": ["user"],
  "client_ip": "192.168.1.100",
  "status": "OK",
  "duration": "15ms"
}
```

### Log Aggregation

For production, send logs to:
- **ELK Stack** (Elasticsearch, Logstash, Kibana)
- **Splunk**
- **Datadog**
- **CloudWatch** (AWS)

---

## 4. Production Checklist

### Security

- [ ] Enable TLS with valid certificates
- [ ] Enable authentication (`auth.enabled: true`)
- [ ] Enable audit logging
- [ ] Configure rate limiting appropriately
- [ ] Use strong client secrets (not defaults)
- [ ] Rotate secrets regularly
- [ ] Enable Keycloak HTTPS
- [ ] Configure firewall rules

### Reliability

- [ ] Set up health check monitoring
- [ ] Configure graceful shutdown timeout
- [ ] Set up log aggregation
- [ ] Configure database backups
- [ ] Set up alerting (Prometheus, Grafana)
- [ ] Test failover scenarios

### Performance

- [ ] Tune rate limits for expected load
- [ ] Configure connection pooling
- [ ] Set appropriate timeouts
- [ ] Enable HTTP/2 (gRPC default)
- [ ] Monitor latency metrics

### Compliance

- [ ] Review audit log retention policy
- [ ] Ensure GDPR compliance (data deletion)
- [ ] Document security controls
- [ ] Perform security audit
- [ ] Set up intrusion detection

---

## 5. Kubernetes Deployment

### Deployment YAML Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
      - name: user-service
        image: your-registry/user-service:latest
        ports:
        - containerPort: 50051
        env:
        - name: APP_AUTH_ENABLED
          value: "true"
        - name: APP_AUTH_KEYCLOAK_URL
          value: "https://keycloak.yourdomain.com"
        - name: APP_SERVER_TLS_CERT
          value: "/certs/tls.crt"
        - name: APP_SERVER_TLS_KEY
          value: "/certs/tls.key"
        volumeMounts:
        - name: tls-certs
          mountPath: /certs
          readOnly: true
        livenessProbe:
          exec:
            command: ["/bin/grpc_health_probe", "-addr=:50051"]
          initialDelaySeconds: 10
        readinessProbe:
          exec:
            command: ["/bin/grpc_health_probe", "-addr=:50051"]
          initialDelaySeconds: 5
      volumes:
      - name: tls-certs
        secret:
          secretName: user-service-tls
```

### Service YAML

```yaml
apiVersion: v1
kind: Service
metadata:
  name: user-service
spec:
  type: LoadBalancer
  ports:
  - port: 50051
    targetPort: 50051
    protocol: TCP
  selector:
    app: user-service
```

---

## 6. Monitoring & Alerting

### Metrics to Monitor

- **Authentication**: Success/failure rate
- **Rate Limiting**: Throttled requests
- **Latency**: p50, p95, p99
- **Error Rate**: 4xx, 5xx responses
- **Throughput**: Requests per second

### Prometheus Integration

Add metrics endpoint to services (future enhancement).

### Alert Examples

```yaml
# High authentication failure rate
- alert: HighAuthFailureRate
  expr: rate(auth_failures[5m]) > 0.1
  annotations:
    summary: "High authentication failure rate detected"

# Rate limit exceeded frequently
- alert: FrequentRateLimiting
  expr: rate(rate_limit_exceeded[5m]) > 10
  annotations:
    summary: "Users hitting rate limits frequently"
```

---

## 7. Disaster Recovery

### Backup Strategy

1. **Database**: Daily automated backups
2. **Keycloak**: Export realm configuration
3. **Certificates**: Secure storage with rotation plan

### Recovery Procedures

1. Restore database from backup
2. Re-import Keycloak realm
3. Deploy services with health checks
4. Verify authentication flow
5. Monitor logs for errors

---

## 8. Security Hardening

### Network Security

- Use VPC/private networks
- Configure security groups
- Enable DDoS protection
- Use Web Application Firewall (WAF)

### Application Security

- Keep dependencies updated
- Run security scans (Snyk, Trivy)
- Implement input validation
- Sanitize error messages
- Use prepared statements (SQL injection prevention)

### Access Control

- Principle of least privilege
- Separate dev/staging/prod environments
- Use service accounts for automation
- Implement MFA for admin access

---

## 9. Performance Tuning

### gRPC Optimizations

```yaml
# Connection pooling
max_connections: 100
max_concurrent_streams: 100

# Timeouts
request_timeout: 30s
keepalive_time: 60s
```

### Rate Limit Tuning

Adjust based on load testing:
```yaml
server:
  rate_limit: 1000  # For high-traffic APIs
```

---

## 10. Compliance

### GDPR

- Implement data deletion endpoints
- Log data access
- Provide data export functionality
- Document data retention policies

### SOC 2

- Enable audit logging
- Implement access controls
- Regular security reviews
- Incident response plan

---

## Support

For production issues:
1. Check audit logs
2. Review health check status
3. Verify Keycloak connectivity
4. Check rate limit metrics
5. Review TLS certificate expiration

**Remember**: Security is a continuous process, not a one-time setup!
