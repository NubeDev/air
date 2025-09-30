#!/bin/bash

echo "🧪 Testing Frontend Slash Command"
echo "================================="

# Start frontend if not running
if ! curl -s http://localhost:3000 > /dev/null; then
    echo "Starting frontend..."
    cd /home/user/code/go/nube/air/air-ui && npm run dev &
    sleep 5
fi

echo "✅ Frontend should be running at http://localhost:3000"
echo "Please test the following in the browser:"
echo "1. Go to http://localhost:3000"
echo "2. Type just '/' in the chat input"
echo "3. Press Enter"
echo "4. You should see the help message with available commands"
echo ""
echo "Expected behavior:"
echo "- Should show available commands (/load, /analyze, /help)"
echo "- Should show current status (loaded dataset, available datasets)"
echo "- Should NOT show 'Unknown command' error"
