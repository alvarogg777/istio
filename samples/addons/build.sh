 helm template -n istio-system prometheus prometheus-community/alertmanager --set persistence.storageClass=longhorn > alert.yaml
 helm template -n istio-system prometheus prometheus-community/kube-state-metrics > metrics.yaml
 helm template -n istio-system prometheus prometheus-community/prometheus-node-exporter > node.yaml
 helm template -n istio-system prometheus prometheus-community/prometheus-pushgateway --set persistentVolume.storageClass=longhorn > pushgateway.yaml
 helm template -n istio-system prometheus prometheus-community/prometheus > pro.yaml
