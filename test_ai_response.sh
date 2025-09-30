#!/bin/bash

echo "🤖 Testing AI Response with Real Data"
echo "====================================="

# Test WebSocket connection and get AI response
node -e "
const WebSocket = require('ws');
const ws = new WebSocket('ws://localhost:9000/v1/ws/');

let step = 0;

ws.on('open', () => {
    console.log('✅ WebSocket connected');
    
    // Step 1: Load dataset
    console.log('📁 Loading dataset...');
    ws.send(JSON.stringify({
        type: 'load_dataset',
        payload: {
            filename: '20250930_200337_air_passengers.csv'
        }
    }));
});

ws.on('message', (data) => {
    try {
        const message = JSON.parse(data);
        
        if (message.type === 'load_dataset_success') {
            console.log('✅ Dataset loaded successfully');
            console.log('📊 Asking AI about headers...');
            
            // Step 2: Ask about headers
            ws.send(JSON.stringify({
                type: 'chat_message',
                payload: {
                    content: 'what is the headers?',
                    model: 'llama'
                }
            }));
        }
        
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

// Timeout after 30 seconds
setTimeout(() => {
    console.log('⏰ Timeout - closing connection');
    ws.close();
    process.exit(1);
}, 30000);
"
