const WebSocket = require('ws');

// Connect to WebSocket
const ws = new WebSocket('ws://localhost:9000/v1/ws');

ws.on('open', function open() {
  console.log('Connected to WebSocket');
  
  // Send file analysis request
  const message = {
    type: 'file_analysis',
    payload: {
      file_id: 'test_file.csv',
      query: 'Analyze this file and show me insights',
      model: 'llama'
    }
  };
  
  console.log('Sending file analysis request:', message);
  ws.send(JSON.stringify(message));
});

ws.on('message', function incoming(data) {
  const message = JSON.parse(data);
  console.log('Received message:', JSON.stringify(message, null, 2));
  
  if (message.type === 'file_analysis_complete') {
    console.log('\n=== File Analysis Complete ===');
    console.log('Analysis:', message.payload.analysis);
    console.log('Insights:', message.payload.insights);
    console.log('Suggestions:', message.payload.suggestions);
    console.log('Data Info:', message.payload.data_info);
    ws.close();
  } else if (message.type === 'file_analysis_error') {
    console.log('Error:', message.payload.error);
    ws.close();
  }
});

ws.on('error', function error(err) {
  console.error('WebSocket error:', err);
});

ws.on('close', function close() {
  console.log('WebSocket connection closed');
});
