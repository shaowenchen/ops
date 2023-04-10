```bash
helm repo add ops https://shaowenchen.github.io/ops
```

```bash
helm install myops ops/ops --version 0.1.0 --namespace ops-system --create-namespace
```