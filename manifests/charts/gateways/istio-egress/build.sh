helm template -nistio-system egress . --set gateways.istio-egressgateway.env.TZ="Europe/Madrid" --set global.tag="1.17.2" --set global.hub="istio" > istio-egress.yaml
