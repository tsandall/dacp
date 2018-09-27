package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/open-policy-agent/opa/loader"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/util"
)

func main() {

	ctx := context.Background()

	// Load Kubernetes resources and policies from disk.
	result, err := loader.All([]string{"."})
	if err != nil {
		panic(err)
	}

	// Load policies into OPA compiler.
	compiler, err := result.Compiler()
	if err != nil {
		panic(err)
	}

	// Load Kubernetes resources into in-memory OPA store.
	store, err := result.Store()
	if err != nil {
		panic(err)
	}

	eval(
		ctx,
		"Example: Query for all violations on all resources.",
		rego.Compiler(compiler),
		rego.Store(store),
		rego.Query(`
			data.example.violation[v]`),
	)

	eval(
		ctx,
		"Example: Query for all 'ingress-host-fqdn' violations on all resources.",
		rego.Compiler(compiler),
		rego.Store(store),
		rego.Query(`
			data.example.violation[{
				"violation": "ingress-host-fqdn",
				"resource": resource,
				"message": message,
			}]`),
	)

	eval(
		ctx,
		"Example: Query for all violations on 'ingress' resources.",
		rego.Compiler(compiler),
		rego.Store(store),
		rego.Query(`
			data.example.violation[{
				"violation": violation,
				"resource": {"kind": "ingress", "namespace": namespace, "name": name},
				"message": message,
			}]`),
	)

	// Example. Temporarily transact ingress and check for violation.
	err = storage.Txn(ctx, store, storage.WriteParams, func(txn storage.Transaction) error {

		// Ingress will be temporarily written into storage at this path. Assumes that parent exists.
		path := storage.MustParsePath("/kubernetes/resources/ingresses/qa/new-ingress")

		// Ingress object to write into storage. The hostname "verify.acmecorp.com"
		// violates the policy because *.acmecorp.com is not whitelisted for the "qa"
		// namespace.
		ingress := util.MustUnmarshalJSON([]byte(`
			{
				"apiVersion": "extensions/v1beta1",
				"kind": "Ingress",
				"metadata": {
					"namespace": "qa",
					"name": "new-ingress"
				},
				"spec": {
					"rules": [
						{
							"host": "verify.acmecorp.com",
							"http": {
								"paths": [
									{
										"backend": {
											"serviceName": "nginx",
											"servicePort": 80
										}
									}
								]
							}
						}
					]
				}
			}
		`))

		// Write the Ingress object into storage.
		err := store.Write(ctx, txn, storage.AddOp, path, ingress)
		if err != nil {
			return err
		}

		// Query for violations for this object. The query is executed within the
		// current transaction to ensure that the ingress written above is visible.
		eval(
			ctx,
			"Example: Query for new violations.",
			rego.Query(`data.example.violation[{
				"violation": violation,
				"resource": {"kind": "ingress", "namespace": "qa", "name": "new-ingress"},
				"message": message,
			}]`),
			rego.Store(store),
			rego.Compiler(compiler),
			rego.Transaction(txn),
		)

		// Boilerplate. Return abort{} error so that storage.Txn aborts the
		// transaction. We ignore the abort{} error below.
		return abort{}
	})

	// Handle unknown errors.
	if err != nil {
		if _, ok := err.(abort); !ok {
			panic(err)
		}
	}
}

type abort struct{}

func (abort) Error() string {
	return "<abort>"
}

func eval(ctx context.Context, note string, options ...func(*rego.Rego)) {

	rs, err := rego.New(options...).Eval(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println("// " + note)

	for _, r := range rs {

		fmt.Println()

		bs, err := json.MarshalIndent(r.Bindings, "", "  ")
		if err != nil {
			panic(err)
		}

		fmt.Println(string(bs))
	}

	fmt.Println()

}
