#!/bin/bash

echo "🔍 Testing Chat Input with Debug Message"
echo "========================================"

# Check if frontend is running
if ! curl -s http://localhost:3000 > /dev/null; then
    echo "❌ Frontend not running. Please start it with 'cd air-ui && npm run dev'"
    exit 1
fi

echo "✅ Frontend is running at http://localhost:3000"
echo ""
echo "Please check the following in the browser:"
echo "1. Go to http://localhost:3000"
echo "2. Look for a RED debug message that says 'DEBUG: Chat input should be visible here'"
echo "3. The chat input should be right below this debug message"
echo ""
echo "If you see the debug message but no chat input:"
echo "- There's an issue with the ChatInput component"
echo "- Check the browser console for errors"
echo ""
echo "If you don't see the debug message at all:"
echo "- The entire input section is not rendering"
echo "- Check if there are any JavaScript errors"
echo ""
echo "The chat input should have:"
echo "- A textarea for typing"
echo "- A paperclip icon for file attachment"
echo "- A send button (arrow icon)"
