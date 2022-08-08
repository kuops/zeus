# zeus

Multi kubernetes cluster query, using informer and listers.

## Dependencies

- gin
- code-generator
- kubebuilder
- swagger

## Quick Start

Generate crd,clientset,informers using code-generator and kubebuilder.

```bash
make update
```

Run the server, default at http://localhost:8080

```bash
make run
```

swagger in http://localhost:8080/swagger/index.html

default generator `admin-cluster` clusters in custom crd `cluster.shiny.io`。

```bash
❯ kubectl get clusters
NAME            VERSION   NODES   PROVIDER
admin-cluster   v1.24.0   1
```
