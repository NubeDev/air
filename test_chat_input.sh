#!/bin/bash

echo "🔍 Testing Chat Input Visibility"
echo "================================"

# Check if frontend is running
if ! curl -s http://localhost:3000 > /dev/null; then
    echo "❌ Frontend not running. Please start it with 'cd air-ui && npm run dev'"
    exit 1
fi

echo "✅ Frontend is running at http://localhost:3000"
echo ""
echo "Please check the following in the browser:"
echo "1. Go to http://localhost:3000"
echo "2. Open Developer Tools (F12)"
echo "3. Check the Console tab for any errors"
echo "4. Look for the chat input at the bottom of the screen"
echo "5. If you don't see it, check the Elements tab and search for 'ChatInput'"
echo ""
echo "The chat input should be visible at the bottom with:"
echo "- A textarea for typing messages"
echo "- A paperclip icon for file attachment"
echo "- A send button (arrow icon)"
echo ""
echo "If the input is missing, there might be a CSS or rendering issue."
