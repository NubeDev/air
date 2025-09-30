#!/bin/bash

echo "🔧 Testing Data Analysis Fix"
echo "============================"

# Check if backend is running
if ! curl -s http://localhost:9000/health > /dev/null; then
    echo "❌ Backend not running. Please start it with 'go run ./cmd/api --data data --config config-dev.yaml --auth disabled'"
    exit 1
fi

echo "✅ Backend is running at http://localhost:9000"
echo ""
echo "🎯 **FIX APPLIED:**"
echo "1. Added more keywords to trigger file analysis:"
echo "   - 'count', 'how many', 'sum', 'total'"
echo "   - 'years', 'data', 'show me', 'find', 'list'"
echo "2. Improved AI prompt to be more specific about data analysis"
echo "3. AI now gets the actual user question + dataset data"
echo ""
echo "Please test the following in the browser:"
echo "1. Go to http://localhost:3000"
echo "2. Load a file (like the air passengers CSV)"
echo "3. Type: 'count all the years'"
echo "4. The AI should now analyze the actual loaded file"
echo "5. Try other queries like:"
echo "   - 'how many years are there?'"
echo "   - 'show me the data'"
echo "   - 'what years are in the dataset?'"
echo ""
echo "The AI should now recognize data analysis requests and use the loaded file!"
