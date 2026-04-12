# GameWalaaBackEnd

## MQTT broker configuration

The backend no longer assumes a local Mosquitto broker. You can point it at a local broker, EMQX Cloud, or any other MQTT-compatible broker by setting these config keys in `config.yml` or environment variables:

```yaml
mqtt_broker_url: "tcp://localhost:1883"
mqtt_client_id: "backend-server"
mqtt_username: ""
mqtt_password: ""
mqtt_ca_cert_path: ""
mqtt_tls_skip_verify: false
```

Environment variable equivalents:

```bash
MQTT_BROKER_URL=tcp://localhost:1883
MQTT_CLIENT_ID=backend-server
MQTT_USERNAME=...
MQTT_PASSWORD=...
MQTT_CA_CERT_PATH=/absolute/path/to/emqx-ca.pem
MQTT_TLS_SKIP_VERIFY=false
```

Supported broker URL schemes:

- `tcp://` for plain MQTT, such as local Mosquitto
- `ssl://` or `tls://` for MQTT over TLS
- `ws://` for MQTT over WebSockets
- `wss://` for MQTT over secure WebSockets

Example for EMQX Cloud over TLS:

```yaml
mqtt_broker_url: "ssl://your-emqx-endpoint:8883"
mqtt_client_id: "backend-server"
mqtt_username: "your-username"
mqtt_password: "your-password"
mqtt_ca_cert_path: "/absolute/path/to/emqx-ca.pem"
mqtt_tls_skip_verify: false
```

Legacy keys `mosquitto_username` and `mosquitto_password` are still accepted as fallbacks while migrating old config files.
