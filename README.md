# Declarative Admission Control Policies for Kubernetes

This repo shows how you can use **the same OPA policy** at enforcement-time
(e.g., in a validating admission controller) as well as offline for audit and
compliance purposes.

## Example Output

``` bash
go run main.go
```

```json
// Example: Query for all violations on all resources.

{
  "v": {
    "message": "invalid ingress host fqdn \"acmecorp.com\"",
    "resource": {
      "kind": "ingress",
      "name": "ingress-bad",
      "namespace": "qa"
    },
    "violation": "ingress-host-fqdn"
  }
}

{
  "v": {
    "message": "ingress host conflicts with an existing ingress",
    "resource": {
      "kind": "ingress",
      "name": "ingress-ok",
      "namespace": "production"
    },
    "violation": "ingress-conflict"
  }
}

{
  "v": {
    "message": "ingress host conflicts with an existing ingress",
    "resource": {
      "kind": "ingress",
      "name": "ingress-conflict",
      "namespace": "staging"
    },
    "violation": "ingress-conflict"
  }
}

// Example: Query for all 'ingress-host-fqdn' violations on all resources.

{
  "message": "invalid ingress host fqdn \"acmecorp.com\"",
  "resource": {
    "kind": "ingress",
    "name": "ingress-bad",
    "namespace": "qa"
  }
}

// Example: Query for all violations on 'ingress' resources.

{
  "message": "invalid ingress host fqdn \"acmecorp.com\"",
  "name": "ingress-bad",
  "namespace": "qa",
  "violation": "ingress-host-fqdn"
}

{
  "message": "ingress host conflicts with an existing ingress",
  "name": "ingress-ok",
  "namespace": "production",
  "violation": "ingress-conflict"
}

{
  "message": "ingress host conflicts with an existing ingress",
  "name": "ingress-conflict",
  "namespace": "staging",
  "violation": "ingress-conflict"
}

// Example: Query for new violations.

{
  "message": "invalid ingress host fqdn \"verify.acmecorp.com\"",
  "violation": "ingress-host-fqdn"
}
```
