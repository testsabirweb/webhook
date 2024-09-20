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

### Output
## nginx
```yaml
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2024-09-20T11:57:30Z"
  labels:
    run: nginx
  name: nginx
  namespace: default
  resourceVersion: "131728"
  uid: a1612e10-4588-481f-9c4c-b392b1ee6aab
spec:
  containers:
  - image: nginx
    imagePullPolicy: Always
    name: nginx
    resources: {}
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: kube-api-access-lnrm4
      readOnly: true
```

## nginx2
```yaml
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2024-09-20T12:02:08Z"
  labels:
    custom_label: custom_value
    spectro: "true"
  name: nginx2
  namespace: default
  resourceVersion: "131957"
  uid: 666e9113-0c36-467f-a186-8c430b7d37d2
spec:
  containers:
  - image: nginx
    imagePullPolicy: Always
    name: nginx2
    resources: {}
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: kube-api-access-s2tpp
      readOnly: true
```
