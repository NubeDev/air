#!/bin/bash

echo "🧪 Testing Slash Command Handling"
echo "================================="

# Test WebSocket connection and slash command
node -e "
const WebSocket = require('ws');
const ws = new WebSocket('ws://localhost:9000/v1/ws/');

ws.on('open', () => {
    console.log('✅ WebSocket connected');
    
    // Test just '/' command
    console.log('Testing "/" command...');
    ws.send(JSON.stringify({
        type: 'chat_message',
        payload: {
            content: '/',
            model: 'llama'
        }
    }));
});

ws.on('message', (data) => {
    try {
        const message = JSON.parse(data);
        
        if (message.type === 'chat_response') {
            console.log('🤖 AI Response:');
            console.log('===============');
            console.log(message.payload.content);
            console.log('===============');
            ws.close();
            process.exit(0);
        }
        
        if (message.type === 'chat_typing') {
            console.log('⏳ AI is thinking...');
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

// Timeout after 10 seconds
setTimeout(() => {
    console.log('⏰ Timeout - closing connection');
    ws.close();
    process.exit(1);
}, 10000);
"
