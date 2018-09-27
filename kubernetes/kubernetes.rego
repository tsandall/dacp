package kubernetes

import data.kubernetes.resources

# Matches provides an abstraction to find resources that match the (kind,
# namespace, name) triplet. In some cases, search logic may be more
# sophisticated.
matches[[kind, namespace, name, resource]] {
    resource := resources[kind][namespace][name]
}