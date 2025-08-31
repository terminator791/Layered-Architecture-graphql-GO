#!/bin/bash

# Example usage script for the Chat Application
# This script demonstrates how to interact with the GraphQL API

echo "🚀 Chat Application Example Usage"
echo "=================================="

# Check if the server is running
echo "Checking if server is running on localhost:8080..."
if ! curl -s http://localhost:8080 > /dev/null; then
    echo "❌ Server is not running. Please start it with:"
    echo "   docker-compose up -d"
    echo "   or"
    echo "   go run cmd/server/main.go"
    exit 1
fi

echo "✅ Server is running!"
echo ""

# Example GraphQL mutations and queries
echo "📝 Example GraphQL Operations:"
echo ""

echo "1. Send a message to the 'general' room:"
echo "mutation {"
echo "  sendMessage(input: {"
echo "    room: \"general\""
echo "    user: \"alice\""
echo "    text: \"Hello, everyone!\""
echo "  }) {"
echo "    id"
echo "    room"
echo "    user"
echo "    text"
echo "    createdAt"
echo "  }"
echo "}"
echo ""

echo "2. Query all messages in the 'general' room:"
echo "query {"
echo "  messages(room: \"general\") {"
echo "    id"
echo "    room"
echo "    user"
echo "    text"
echo "    createdAt"
echo "  }"
echo "}"
echo ""

echo "3. Subscribe to real-time messages (WebSocket):"
echo "subscription {"
echo "  messageAdded(room: \"general\") {"
echo "    id"
echo "    room"
echo "    user"
echo "    text"
echo "    createdAt"
echo "  }"
echo "}"
echo ""

echo "🌐 Access the GraphQL Playground at: http://localhost:8080"
echo "🔗 GraphQL endpoint: http://localhost:8080/query"
echo ""

echo "💡 To test real-time functionality:"
echo "   1. Open two browser tabs with the GraphQL Playground"
echo "   2. In the first tab, set up the subscription above"
echo "   3. In the second tab, send a message using the mutation"
echo "   4. Watch the message appear in real-time in the first tab!"