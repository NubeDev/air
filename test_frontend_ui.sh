#!/bin/bash

echo "🧪 Testing Frontend UI Improvements"
echo "===================================="

# Start frontend if not running
if ! curl -s http://localhost:3000 > /dev/null; then
    echo "Starting frontend..."
    cd /home/user/code/go/nube/air/air-ui && npm run dev &
    sleep 5
fi

echo "✅ Frontend should be running at http://localhost:3000"
echo ""
echo "Please test the following in the browser:"
echo "1. Go to http://localhost:3000"
echo "2. Type '/load' in the chat input"
echo "3. Press Enter"
echo ""
echo "Expected behavior:"
echo "- Should show numbered list of available datasets"
echo "- Should show usage: /load <filename> or /load <number>"
echo "- Should show examples with actual filenames"
echo "- Should NOT show generic AI response"
echo ""
echo "If you see a generic AI response, the slash command logic needs fixing."
