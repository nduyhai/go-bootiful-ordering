#!/bin/bash

# Wait for Debezium Connect to be ready
echo "Waiting for Debezium Connect to be ready..."
until curl -s http://debezium:8083/ > /dev/null; do
  sleep 5
done
echo "Debezium Connect is ready."

# Register the connector
echo "Registering Debezium connector for order outbox..."
curl -i -X POST -H "Accept:application/json" -H "Content-Type:application/json" \
  http://debezium:8083/connectors/ -d @/app/debezium-connector-config.json

echo "Connector registration completed."