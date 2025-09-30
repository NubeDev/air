#!/bin/bash

echo "🔌 Testing WebSocket Connection Fix"
echo "==================================="

# Check if frontend is running
if ! curl -s http://localhost:3000 > /dev/null; then
    echo "❌ Frontend not running. Please start it with 'cd air-ui && npm run dev'"
    exit 1
fi

echo "✅ Frontend is running at http://localhost:3000"
echo ""
echo "🎯 **FIX APPLIED:**"
echo "1. Added WebSocket connection initialization to ChatWindowNew"
echo "2. WebSocket will now connect when the chat component loads"
echo ""
echo "Please check the following in the browser:"
echo "1. Go to http://localhost:3000"
echo "2. Open Developer Tools (F12)"
echo "3. Check the Console tab"
echo "4. You should see: 'WebSocket connected in ChatWindow'"
echo "5. Try sending a message to test the connection"
echo ""
echo "If you still see 'WebSocket not connected' errors:"
echo "- Check the browser console for connection errors"
echo "- Try refreshing the page"
echo "- Make sure the backend is running on port 9000"
echo ""
echo "The WebSocket should now connect automatically when you load the chat page!"
