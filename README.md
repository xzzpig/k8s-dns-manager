# Kubernetes DNS Manager

### ⚠️ Project Deprecated ⚠️

**Warning: This project is deprecated. Please use [kube-dns-manager](https://github.com/xzzpig/kube-dns-manager) instead.**

This project is no longer maintained. All features and updates have been migrated to the new project [kube-dns-manager](https://github.com/xzzpig/kube-dns-manager). Please switch to the new project as soon as possible to receive the latest features and support.

Thank you for your understanding and support!

## Overview
### DNSRecord
> allows you to manage DNS records for your Kubernetes cluster using custom resources.

Example A Record:
```yaml
apiVersion: dns.xzzpig.com/v1
kind: DNSRecord
metadata:
  name: dnsrecord-sample
  namespace: default
spec:
  recordType: A
  name: test.sample.com
  value: 192.168.1.1
```

### DNSProvider
> you can use this resource to configure the DNS provider and credentials to use. The `k8s-dns-manager` will match `DNSRecord` with ***one*** `DNSProvider` and sync the DNS records in the configured DNS provider. Specially, `DNSProvider` is cluster-scoped.

Example DNSProvider:
```yaml
apiVersion: dns.xzzpig.com/v1
kind: DNSProvider
metadata:
  name: dnsprovider-sample
spec:
  providerType: ALIYUN
  domainName: sample.com
  aliyun:
    accessKeyId: "<your-access-key-id>"
    accessKeySecret: "<your-access-key-secret>"
```


### Auto Generate DNS Records
#### Ingress
> `k8s-dns-manager` will create an A `DNSRecord` per host with the announced `DNSGenerator`

For example
> create a generator first
```yaml
apiVersion: dns.xzzpig.com/v1
kind: DNSGenerator
metadata:
  name: ddns
spec:
  generatorType: DDNS
```
> add annotation `dns.xzzpig.com/generator` to the ingress to use the generator
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test
  annotations:
    dns.xzzpig.com/generator: ddns
    dns.xzzpig.com/record-proxied: "true" #Annotation/Label start with `dns.xzzpig.com/record-` will be copied to the DNSRecord
spec:
  rules:
  - host: test.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: test-service
            port:
              number: 80
```

## Installation
```bash
kubectl apply -f https://github.com/xzzpig/k8s-dns-manager/raw/main/deploy/manifests.yaml
```

## Uninstall
```bash
kubectl delete -f https://github.com/xzzpig/k8s-dns-manager/raw/main/deploy/bundle.yaml  
```

## Model
![Model](model.drawio.svg)

## Reference
### Supported DNS Providers
#### Aliyun
```yaml
apiVersion: dns.xzzpig.com/v1
kind: DNSProvider
metadata:
  name: dnsprovider-sample-aliyun
spec:
  providerType: ALIYUN
  domainName: sample.com
  aliyun:
    accessKeyId: "<your-access-key-id>"
    accessKeySecret: "<your-access-key-secret>"
```

#### Cloudflare
```yaml
apiVersion: dns.xzzpig.com/v1
kind: DNSProvider
metadata:
  name: dnsprovider-sample-cloudflare
spec:
  providerType: CLOUDFLARE
  domainName: sample.com
  cloudflare:
    proxied: false #If true, the DNS record will be proxied by Cloudflare, can be overrided by Annotation `dns.xzzpig.com/record-proxied`, default is false
    zoneName: sample.com # If empty, spec.domainName will be used as zone name
    #use apiToken or key+email to auth
    #if both are set, apiToken will be used
    apiToken: "<your-api-token>"
    key: "<your-key>"
    email: "<your-email>"
```

### Supported DNS Types
- A
- CNAME
- TXT
- MX
- SRV
- AAAA
- NS
- CAA

## Environment Variables
| Name | Description | Type | Default |
| --- | --- | --- | --- |
| GO_ENV | The environment of the application | string | `production` |
| NATM_DEFAULT_RECORD_TTL | The default TTL for DNS records | int | `600` |
| NATM_DEFAULT_GENERATOR_TYPE | The default generator type for DNS records, will be used when auto generate dns record if the generator type is not specified, will be ignored when the value is empty string | string |  |
| NATM_BIND_METRICS | The address to bind the metrics server | string | `:8080` |
| NATM_BIND_HEALTH_PROBE | The address to bind the health probe server | string | `:8081` |

## Generate DNS Records
### Supported Targets
- Ingress
### Supported `DNSGenerator` Types
| Type | Description | Support Target |
| --- | --- | --- |
| DDNS | Type is A; Value is the public ip of the ingress controller get by ddns service | Ingress |
| CNAME | Type is CNAME; Value can be overwrited by annotation `dns.xzzpig.com/cname` on the target | Ingress |

## Annotations
| Name | Description | Target |
| --- | --- | --- |
| dns.xzzpig.com/generator | The `Generator Types` for DNS records | Ingress |
| dns.xzzpig.com/cname | The value of CNAME record | Ingress(`generator`=`cname`) |
| dns.xzzpig.com/record-proxied | `DNSRecord` will be set as proxied  | Ingress DNSRecord(`recordType`=`CLOUDFLARE`) |

## TODO
- [ ] Support more DNS providers
    - [x] Aliyun
    - [x] Cloudflare
    - [ ] DNSPod
- [ ] Auto generate DNS records for more targets
    - [x] Ingress
    - [ ] Service
