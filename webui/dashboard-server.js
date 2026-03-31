const express = require('express');
const fs = require('fs');
const path = require('path');
const cors = require('cors');

const app = express();
const PORT = 4141;

app.use(cors());
app.use(express.json());

// API to get list of all test result files
app.get('/api/test-runs', (req, res) => {
  const testResultsDir = path.join(__dirname, 'test-results');  // ← REMOVE ../
  
  //console.log('Looking for test results in:', testResultsDir); // ← Add debug log
  
  try {
    if (!fs.existsSync(testResultsDir)) {
      console.log('❌ test-results directory does not exist!');
      return res.json([]);
    }

    const files = fs.readdirSync(testResultsDir)
      .filter(file => file.endsWith('.json'))
      .map(file => {
        const filePath = path.join(testResultsDir, file);
        const stats = fs.statSync(filePath);
        
        // Extract timestamp from filename
        const timestamp = file.match(/results-(\d+)\.json/)?.[1];
        
        return {
          id: file,
          filename: file,
          timestamp: timestamp ? parseInt(timestamp) : stats.mtimeMs,
          date: new Date(timestamp ? parseInt(timestamp) : stats.mtimeMs).toLocaleString(),
          size: stats.size
        };
      })
      .sort((a, b) => b.timestamp - a.timestamp);
    
    //console.log(`✅ Found ${files.length} test result files`);
    res.json(files);
  } catch (error) {
    console.error('Error reading test results:', error);
    res.status(500).json({ error: 'Failed to read test results' });
  }
});

// API to get specific test result details
app.get('/api/test-runs/:filename', (req, res) => {
  const testResultsDir = path.join(__dirname, 'test-results');  // ← FIXED
  const filePath = path.join(testResultsDir, req.params.filename);  // ← FIXED
  
  //console.log('Reading file:', filePath); // ← Add debug log
  
  try {
    const data = fs.readFileSync(filePath, 'utf8');
    res.json(JSON.parse(data));
  } catch (error) {
    console.error('Error reading test result:', error);
    res.status(404).json({ error: 'Test result not found' });
  }
});

app.listen(PORT, () => {
  console.log(`🚀 Test results server running on http://localhost:${PORT}`);
  console.log(`📁 Looking for test results in: ${path.join(__dirname, 'test-results')}`);
});
