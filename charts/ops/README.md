```bash
helm repo add ops https://www.chenshaowen.com/ops/charts
```

```bash
helm install myops ops/ops --version 2.0.0 --namespace ops-system --create-namespace
```