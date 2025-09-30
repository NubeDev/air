#!/bin/bash

echo "🧪 Testing AIR File Analysis Workflow"
echo "====================================="

# Check if backend is running
echo "1. Checking if backend is running..."
if ! curl -s http://localhost:9000/v1/health > /dev/null; then
    echo "❌ Backend not running. Starting it..."
    cd /home/user/code/go/nube/air
    make clean-ports
    make dev-backend
    sleep 5
else
    echo "✅ Backend is running"
fi

# Check if frontend is running
echo "2. Checking if frontend is running..."
if ! curl -s http://localhost:5173 > /dev/null; then
    echo "❌ Frontend not running. Starting it..."
    cd /home/user/code/go/nube/air/air-ui
    npm run dev &
    sleep 5
else
    echo "✅ Frontend is running"
fi

# Test WebSocket connection
echo "3. Testing WebSocket connection..."
node -e "
const WebSocket = require('ws');
const ws = new WebSocket('ws://localhost:9000/v1/ws/');

ws.on('open', () => {
    console.log('✅ WebSocket connected');
    
    // Test load dataset
    console.log('4. Testing load dataset...');
    ws.send(JSON.stringify({
        type: 'load_dataset',
        payload: {
            filename: '20250930_200337_air_passengers.csv'
        }
    }));
    
    // Wait a bit then test chat
    setTimeout(() => {
        console.log('5. Testing chat with loaded file...');
        ws.send(JSON.stringify({
            type: 'chat_message',
            payload: {
                content: 'what is the headers?',
                model: 'llama'
            }
        }));
    }, 2000);
    
    // Wait for responses
    setTimeout(() => {
        console.log('6. Closing connection...');
        ws.close();
        process.exit(0);
    }, 15000);
});

ws.on('message', (data) => {
    try {
        const message = JSON.parse(data);
        console.log('📨 Received:', message.type);
        if (message.payload && message.payload.content) {
            console.log('   Content:', message.payload.content.substring(0, 100) + '...');
        }
    } catch (e) {
        console.log('📨 Raw data:', data.toString().substring(0, 100) + '...');
    }
});

ws.on('error', (error) => {
    console.log('❌ WebSocket error:', error.message);
});

ws.on('close', () => {
    console.log('🔌 WebSocket closed');
});
"

echo "7. Test completed!"