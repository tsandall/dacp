package example

import data.kubernetes.matches
import data.kubernetes.cluster_resources.namespaces
import data.kubernetes.resources.ingresses

##############################################################################
#
# Policy 1: Ingress hostnames must be whitelisted on namespace.
#
# This policy shows how you can leverage context beyond an individual resource
# to make decisions. In this case, the whitelist is stored on the Namespace
# associated with the Ingress. To decide whether the Ingress hostname violates
# our policy, we check if it matches of the whitelisted patterns stored on the
# Namespace.
#
##############################################################################

violation[{
    "violation": "ingress-host-fqdn",   # identifies type of violation
    "resource": {
        "kind": "ingress",              # identifies kind of resource
        "namespace": namespace,         # identifies namespace of resource
        "name": name                    # identifies name of resource
    },
    "message": msg,                     # provides human-readable message to display
}] {
    matches[["ingresses", namespace, name, ingress]]
    host := ingress.spec.rules[_].host
    valid_hosts := valid_ingress_hosts(namespace)
    not fqdn_matches_any(host, valid_hosts)
    msg := sprintf("invalid ingress host fqdn %q", [host])
}

valid_ingress_hosts(namespace) = {host |
    whitelist := namespaces[namespace].metadata.annotations["ingress-whitelist"]
    hosts := split(whitelist, ",")
    host := hosts[_]
}

fqdn_matches_any(str, patterns) {
    fqdn_matches(str, patterns[_])
}

fqdn_matches(str, pattern) {
    pattern_parts = split(pattern, ".")
    pattern_parts[0] = "*"
    str_parts = split(str, ".")
    n_pattern_parts = count(pattern_parts)
    n_str_parts = count(str_parts)
    suffix = trim(pattern, "*.")
    endswith(str, suffix)
}

fqdn_matches(str, pattern) {
    not contains(pattern, "*")
    str == pattern
}

##############################################################################
#
# Policy 2: Ingress hostnames must be unique across Namespaces.
#
# This policy shows how you can express a pair-wise search. In this case, there
# is a violation if any two ingresses in different namespaces. Note, you can
# query OPA to determine whether a single Ingress violates the policy (in which
# case the cost is linear with the # of Ingresses) or you can query for the set
# of all Ingresses th violate the policy (in which case the cost is (# of
# Ingresses)^2.)
#
##############################################################################

violation[{
    "violation": "ingress-conflict",
    "resource": {"kind": "ingress", "namespace": namespace, "name": name},
    "message": "ingress host conflicts with an existing ingress",
}] {
    matches[["ingresses", namespace, name, ingress]]
    matches[["ingresses", other_ns, other_name, other_ingress]]
    namespace != other_ns
    other_ingress.spec.rules[_].host == ingress.spec.rules[_].host
}
