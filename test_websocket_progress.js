const WebSocket = require('ws');

// Connect to WebSocket
const ws = new WebSocket('ws://localhost:9000/v1/ws/');

ws.on('open', function open() {
  console.log('✅ Connected to WebSocket');
  
  // Send file analysis request
  const message = {
    type: 'file_analysis',
    payload: {
      file_id: '20250930_201837_test_file.csv', // Use the uploaded file
      query: 'Analyze this file and show me insights',
      model: 'llama'
    }
  };
  
  console.log('📤 Sending file analysis request:', JSON.stringify(message, null, 2));
  ws.send(JSON.stringify(message));
});

ws.on('message', function incoming(data) {
  // Handle multiple messages that might be concatenated
  const messages = data.toString().split('\n').filter(line => line.trim());
  
  messages.forEach(messageStr => {
    try {
      const message = JSON.parse(messageStr);
      console.log('📥 Received message:', JSON.stringify(message, null, 2));
  
      if (message.type === 'file_analysis_started') {
        console.log('🚀 Analysis started!');
      } else if (message.type === 'file_analysis_progress') {
        console.log(`🔄 Progress: ${message.payload.message} (${message.payload.progress}%)`);
      } else if (message.type === 'file_analysis_complete') {
        console.log('\n✅ === File Analysis Complete ===');
        console.log('📊 Analysis:', message.payload.analysis);
        console.log('💡 Insights:', message.payload.insights);
        console.log('🎯 Suggestions:', message.payload.suggestions);
        console.log('📈 Data Info:', message.payload.data_info);
        ws.close();
      } else if (message.type === 'file_analysis_error') {
        console.log('❌ Error:', message.payload.error);
        ws.close();
      }
    } catch (error) {
      console.error('❌ Failed to parse message:', messageStr, error.message);
    }
  });
});

ws.on('error', function error(err) {
  console.error('❌ WebSocket error:', err);
});

ws.on('close', function close() {
  console.log('🔌 WebSocket connection closed');
});
