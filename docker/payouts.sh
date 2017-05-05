#!/bin/bash

# Wait until PostgreSQL started and listens on port 5432.
while [ -z "`netstat -tln | grep 8545`" ]; do
  echo 'Waiting for gmc to start ...'
  sleep 1
done
echo 'gmc started.'
sleep 5

# Start server.
echo 'Starting payouts...'
/opt/open-ethereum-pool/build/bin/open-ethereum-pool /opt/open-ethereum-pool/payouts.json
