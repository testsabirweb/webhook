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

### Webhook 
#### Mutation Webhook Controller
```bash
kubectl apply -f manifests/webhook-deploy.yaml
```

#### Webhook
```bash
CA_CERT=$(cat tls.crt | base64 | tr -d '\n')
sed -e 's@CA-CERT@'"$CA_CERT"'@g' <"manifests/webhook-template.yaml" > manifests/webhook.yaml
kubectl apply -f manifests/webhook.yaml
```

### Testing
```bash
kubectl run nginx --image=nginx --restart=Never
kubectl run nginx2 --image=nginx --restart=Never --labels="spectro=true"
```

```bash

kk apply -f manifests/configmap.yaml
kk apply -f manifests/service-account.yaml
kk apply -f manifests/role.yaml
kk apply -f manifests/role-binding.yaml


kk apply -f manifests/webhook-deploy.yaml 

kk apply -f manifests/webhook.yaml 

kk run nginx --image=nginx --restart=Never -n sabir
```

```bash
kk describe pod/nginx
```

```yaml
Name:             nginx
Namespace:        default
Priority:         0
Service Account:  default
Node:             edge-ec2b652c1eb48dd2acf82802b05d1f12/10.0.2.158
Start Time:       Tue, 24 Sep 2024 09:55:14 +0000
Labels:           environment=production
                  run=nginx
                  team=devops
```