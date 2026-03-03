#!/bin/bash

SERVER_URL="${SERVER_URL:-http://localhost:8086}"

echo "Testing Messaging API"
echo "====================="
echo ""

echo "1. Health Check"
echo "---------------"
curl -s "${SERVER_URL}/health" | jq .
echo -e "\n"

echo "2. Sending test message"
echo "-----------------------"
curl -s -X POST "${SERVER_URL}/api/messages" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Test message from script",
    "metadata": {
      "source": "test-script",
      "type": "automated"
    }
  }' | jq .
echo -e "\n"

echo "3. Sending multiple messages"
echo "----------------------------"
for i in {1..5}; do
  echo "Sending message $i..."
  curl -s -X POST "${SERVER_URL}/api/messages" \
    -H "Content-Type: application/json" \
    -d "{
      \"content\": \"Message number $i\",
      \"metadata\": {
        \"source\": \"batch-test\",
        \"type\": \"load-test\"
      }
    }" | jq -r '.message'
done

echo -e "\nTest completed!"
