# Resource Webhook

### TLS Certs
#### CA Certificates
```bash
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 365 -key ca.key -subj "/C=CN/ST=GD/L=SZ/O=Acme, Inc./CN=Acme Root CA" -out ca.crt
```


#### Issue TLS certificates
TLS certifcates for `spectro-webhook` service 
```bash
export SERVICE=spectro-webhook
openssl req -newkey rsa:2048 -nodes -keyout tls.key -subj "/C=CN/ST=GD/L=SZ/O=Acme, Inc./CN=$SERVICE.default.svc.cluster.local" -out tls.csr
openssl x509 -req -extfile <(printf "subjectAltName=DNS:$SERVICE.default.svc.cluster.local,DNS:$SERVICE.default.svc.cluster,DNS:$SERVICE.default.svc,DNS:$SERVICE.default.svc,DNS:$SERVICE.default,DNS:$SERVICE") -days 365 -in tls.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out tls.crt

sudo cp tls.crt tls.key /etc/certs
```

#### Create TLS kubernetes Secret
```bash
kubectl create secret tls tls --cert=tls.crt --key=tls.key
```

#### Webhook
```bash
CA_CERT=$(cat tls.crt | base64 | tr -d '\n')
sed -e 's@CA-CERT@'"$CA_CERT"'@g' <"manifests/webhook-template.yaml" > manifests/webhook.yaml
kubectl apply -f manifests/webhook.yaml
```
