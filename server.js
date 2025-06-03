const express = require('express');
const path = require('path');
const fs = require('fs');

const app = express();
const PORT = 3000;

// Serve static files (HTML, CSS, JS)
app.use(express.static('.'));

// Function to read .env file
function readEnvFile() {
    try {
        const envPath = path.join(__dirname, '.env');
        
        if (!fs.existsSync(envPath)) {
            return null;
        }
        
        const envContent = fs.readFileSync(envPath, 'utf8');
        const envVars = {};
        
        envContent.split('\n').forEach(line => {
            line = line.trim();
            if (line && !line.startsWith('#')) {
                const [key, ...valueParts] = line.split('=');
                if (key && valueParts.length > 0) {
                    envVars[key.trim()] = valueParts.join('=').trim();
                }
            }
        });
        
        return envVars;
    } catch (error) {
        console.error('Error reading .env file:', error);
        return null;
    }
}

// API endpoint to get the Readwise token
app.get('/api/token', (req, res) => {
    const envVars = readEnvFile();
    
    if (!envVars || !envVars.READWISE_TOKEN) {
        return res.status(404).json({ 
            error: 'READWISE_TOKEN not found in .env file',
            message: 'Please create a .env file with READWISE_TOKEN=your_token_here'
        });
    }
    
    res.json({ token: envVars.READWISE_TOKEN });
});

// API endpoint to get environment info
app.get('/api/env-info', (req, res) => {
    const envVars = readEnvFile();
    
    if (!envVars) {
        return res.json({
            hasEnvFile: false,
            hasToken: false,
            message: 'No .env file found'
        });
    }
    
    res.json({
        hasEnvFile: true,
        hasToken: !!envVars.READWISE_TOKEN,
        message: envVars.READWISE_TOKEN ? 'Token found in .env file' : 'READWISE_TOKEN not found in .env file'
    });
});

// Serve the main page
app.get('/', (req, res) => {
    res.sendFile(path.join(__dirname, 'index.html'));
});

app.listen(PORT, () => {
    console.log(`Daily Tech Digest server running at http://localhost:${PORT}`);
    console.log('Open your browser and navigate to the URL above');
    
    // Check if .env file exists on startup
    const envVars = readEnvFile();
    if (envVars && envVars.READWISE_TOKEN) {
        console.log('✅ Found READWISE_TOKEN in .env file');
    } else {
        console.log('⚠️  No .env file or READWISE_TOKEN found');
        console.log('   Create a .env file with: READWISE_TOKEN=your_token_here');
    }
}); 