```bash
helm repo add ops https://shaowenchen.github.io/ops/charts
```

```bash
helm install myops ops/ops --version 1.0.0 --namespace ops-system --create-namespace
```