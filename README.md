# Kubernetes DNS Manager

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
> `k8s-dns-manager` will create an A `DNSRecord` per host with the announced `Generator Types`

For example
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
| NATM_DEFAULT_GENERATOR_DDNS_TIMEOUT | The timeout for ddns service for generator type `ddns` | Duration | `2s` |
| NATM_DEFAULT_GENERATOR_DDNS_EXTRA_APIS | The extra apis to get public ip for generator type `ddns` | []string |  |
| NATM_DEFAULT_GENERATOR_DDNS_CACHE_EXPIRE | The expire time for public ip cache for generator type `ddns` | Duration | `1m` |
| NATM_DEFAULT_GENERATOR_DDNS_CLEAN_INTERVAL | The interval to clean the public ip cache for generator type `ddns` | Duration | `30s` |
| NATM_DEFAULT_GENERATOR_DDNS_REFRESH_INTERNAL | The interval to refresh the public ip for generator type `ddns` | Duration | `10m` |
| NATM_DEFAULT_GENERATOR_CNAME_VALUE | The default value for generator type `cname` | string |  |

## Generate DNS Records
### Supported Targets
- Ingress
### Supported `Generator Types`
| Name | Description | Target |
| --- | --- | --- |
| ddns | Type is A; Value is the public ip of the ingress controller get by ddns service | Ingress |
| cname | Type is CNAME; Value should be annotated by `dns.xzzpig.com/cname` on the target | Ingress |

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
