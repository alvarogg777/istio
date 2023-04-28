helm template -nistio-system egress . --set pilot.env.env.TZ="Europe/Madrid" --set pilot.tag="1.17.2" --set pilot.hub="istio" --set global.tag="1.17.2" > discovery.yaml
