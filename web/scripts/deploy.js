#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const StaticSiteGenerator = require('./build');

class DeploymentManager {
    constructor() {
        this.buildDir = path.join(__dirname, '..', 'static');
        this.deployConfig = this.loadDeployConfig();
    }

    loadDeployConfig() {
        const configPath = path.join(__dirname, '..', 'deploy-config.json');
        
        const defaultConfig = {
            target: process.env.DEPLOY_TARGET || 'local',
            s3: {
                bucket: process.env.S3_BUCKET || 'autonomous-content-web',
                region: process.env.AWS_REGION || 'us-east-1',
                cloudfront: process.env.CLOUDFRONT_DISTRIBUTION_ID || null
            },
            ftp: {
                host: process.env.FTP_HOST || null,
                username: process.env.FTP_USERNAME || null,
                password: process.env.FTP_PASSWORD || null,
                path: process.env.FTP_PATH || '/'
            },
            git: {
                repository: process.env.GIT_DEPLOY_REPO || null,
                branch: process.env.GIT_DEPLOY_BRANCH || 'gh-pages'
            }
        };

        if (fs.existsSync(configPath)) {
            try {
                const fileConfig = JSON.parse(fs.readFileSync(configPath, 'utf8'));
                return { ...defaultConfig, ...fileConfig };
            } catch (error) {
                console.warn('Error loading deploy config, using defaults:', error.message);
            }
        }

        return defaultConfig;
    }

    async deploy() {
        console.log('Starting deployment process...');
        
        try {
            // Build the site first
            const generator = new StaticSiteGenerator();
            await generator.build();
            
            // Deploy based on target
            switch (this.deployConfig.target) {
                case 's3':
                    await this.deployToS3();
                    break;
                case 'ftp':
                    await this.deployToFTP();
                    break;
                case 'git':
                    await this.deployToGit();
                    break;
                case 'local':
                default:
                    await this.deployLocal();
                    break;
            }
            
            console.log('Deployment completed successfully!');
            
        } catch (error) {
            console.error('Deployment failed:', error);
            process.exit(1);
        }
    }

    async deployLocal() {
        console.log('Local deployment - files are ready in ./static directory');
        console.log(`Site available at: file://${this.buildDir}/index.html`);
        
        // Optional: Start a local server
        if (process.argv.includes('--serve')) {
            this.startLocalServer();
        }
    }

    async deployToS3() {
        console.log('Deploying to AWS S3...');
        
        try {
            const AWS = require('aws-sdk');
            const s3 = new AWS.S3({ region: this.deployConfig.s3.region });
            
            await this.uploadDirectoryToS3(s3, this.buildDir);
            
            if (this.deployConfig.s3.cloudfront) {
                await this.invalidateCloudFront();
            }
            
            console.log(`Deployed to S3: https://${this.deployConfig.s3.bucket}.s3.amazonaws.com`);
            
        } catch (error) {
            if (error.code === 'MODULE_NOT_FOUND') {
                console.error('AWS SDK not found. Install with: npm install aws-sdk');
            } else {
                console.error('S3 deployment failed:', error);
            }
            throw error;
        }
    }

    async uploadDirectoryToS3(s3, dirPath) {
        const files = this.getAllFiles(dirPath);
        
        for (const file of files) {
            const relativePath = path.relative(dirPath, file);
            const fileContent = fs.readFileSync(file);
            const contentType = this.getContentType(path.extname(file));
            
            const params = {
                Bucket: this.deployConfig.s3.bucket,
                Key: relativePath.replace(/\\/g, '/'), // Ensure forward slashes for S3
                Body: fileContent,
                ContentType: contentType,
                ACL: 'public-read'
            };
            
            await s3.upload(params).promise();
            console.log(`Uploaded: ${relativePath}`);
        }
    }

    async invalidateCloudFront() {
        try {
            const AWS = require('aws-sdk');
            const cloudfront = new AWS.CloudFront();
            
            const params = {
                DistributionId: this.deployConfig.s3.cloudfront,
                InvalidationBatch: {
                    CallerReference: Date.now().toString(),
                    Paths: {
                        Quantity: 1,
                        Items: ['/*']
                    }
                }
            };
            
            await cloudfront.createInvalidation(params).promise();
            console.log('CloudFront invalidation created');
            
        } catch (error) {
            console.warn('CloudFront invalidation failed:', error.message);
        }
    }

    async deployToFTP() {
        console.log('Deploying via FTP...');
        
        try {
            const ftpClient = require('ftp');
            const client = new ftpClient();
            
            await new Promise((resolve, reject) => {
                client.connect({
                    host: this.deployConfig.ftp.host,
                    user: this.deployConfig.ftp.username,
                    password: this.deployConfig.ftp.password
                });
                
                client.on('ready', () => {
                    this.uploadDirectoryToFTP(client, this.buildDir, this.deployConfig.ftp.path)
                        .then(resolve)
                        .catch(reject);
                });
                
                client.on('error', reject);
            });
            
            console.log('FTP deployment completed');
            
        } catch (error) {
            if (error.code === 'MODULE_NOT_FOUND') {
                console.error('FTP module not found. Install with: npm install ftp');
            } else {
                console.error('FTP deployment failed:', error);
            }
            throw error;
        }
    }

    async deployToGit() {
        console.log('Deploying to Git Pages...');
        
        try {
            const { execSync } = require('child_process');
            const tempDir = path.join(__dirname, '..', '.deploy-temp');
            
            // Clone or update repository
            if (!fs.existsSync(tempDir)) {
                execSync(`git clone ${this.deployConfig.git.repository} ${tempDir}`);
            } else {
                execSync('git pull', { cwd: tempDir });
            }
            
            // Switch to deploy branch
            try {
                execSync(`git checkout ${this.deployConfig.git.branch}`, { cwd: tempDir });
            } catch {
                execSync(`git checkout -b ${this.deployConfig.git.branch}`, { cwd: tempDir });
            }
            
            // Copy built files
            this.copyDirectory(this.buildDir, tempDir);
            
            // Commit and push
            execSync('git add .', { cwd: tempDir });
            execSync(`git commit -m "Deploy: ${new Date().toISOString()}"`, { cwd: tempDir });
            execSync(`git push origin ${this.deployConfig.git.branch}`, { cwd: tempDir });
            
            console.log('Git deployment completed');
            
        } catch (error) {
            console.error('Git deployment failed:', error);
            throw error;
        }
    }

    startLocalServer() {
        const http = require('http');
        const url = require('url');
        const port = process.env.PORT || 8080;
        
        const server = http.createServer((req, res) => {
            const pathname = url.parse(req.url).pathname;
            let filePath = path.join(this.buildDir, pathname);
            
            // Default to index.html for directory requests
            if (pathname === '/' || pathname.endsWith('/')) {
                filePath = path.join(filePath, 'index.html');
            }
            
            // Add .html extension if file doesn't exist
            if (!fs.existsSync(filePath) && !path.extname(filePath)) {
                filePath += '.html';
            }
            
            if (fs.existsSync(filePath)) {
                const ext = path.extname(filePath);
                const contentType = this.getContentType(ext);
                
                res.writeHead(200, { 'Content-Type': contentType });
                res.end(fs.readFileSync(filePath));
            } else {
                res.writeHead(404, { 'Content-Type': 'text/html' });
                res.end('<h1>404 - Page Not Found</h1>');
            }
        });
        
        server.listen(port, () => {
            console.log(`Local server running at http://localhost:${port}`);
        });
    }

    getAllFiles(dirPath, arrayOfFiles = []) {
        const files = fs.readdirSync(dirPath);
        
        files.forEach((file) => {
            const fullPath = path.join(dirPath, file);
            if (fs.statSync(fullPath).isDirectory()) {
                arrayOfFiles = this.getAllFiles(fullPath, arrayOfFiles);
            } else {
                arrayOfFiles.push(fullPath);
            }
        });
        
        return arrayOfFiles;
    }

    copyDirectory(src, dest) {
        if (!fs.existsSync(dest)) {
            fs.mkdirSync(dest, { recursive: true });
        }
        
        const items = fs.readdirSync(src);
        items.forEach(item => {
            const srcPath = path.join(src, item);
            const destPath = path.join(dest, item);
            
            if (fs.statSync(srcPath).isDirectory()) {
                this.copyDirectory(srcPath, destPath);
            } else {
                fs.copyFileSync(srcPath, destPath);
            }
        });
    }

    getContentType(ext) {
        const types = {
            '.html': 'text/html',
            '.css': 'text/css',
            '.js': 'application/javascript',
            '.json': 'application/json',
            '.png': 'image/png',
            '.jpg': 'image/jpeg',
            '.jpeg': 'image/jpeg',
            '.gif': 'image/gif',
            '.svg': 'image/svg+xml',
            '.ico': 'image/x-icon',
            '.woff': 'font/woff',
            '.woff2': 'font/woff2',
            '.ttf': 'font/ttf',
            '.xml': 'application/xml',
            '.txt': 'text/plain'
        };
        
        return types[ext.toLowerCase()] || 'application/octet-stream';
    }
}

if (require.main === module) {
    const deployer = new DeploymentManager();
    
    const args = process.argv.slice(2);
    if (args.includes('--help')) {
        console.log(`
Usage: node deploy.js [options]

Options:
  --serve     Start local server after deployment
  --help      Show this help message

Environment Variables:
  DEPLOY_TARGET         Deployment target (local, s3, ftp, git)
  S3_BUCKET            S3 bucket name
  AWS_REGION           AWS region
  CLOUDFRONT_DISTRIBUTION_ID  CloudFront distribution ID
  FTP_HOST             FTP server hostname
  FTP_USERNAME         FTP username
  FTP_PASSWORD         FTP password
  FTP_PATH             FTP remote path
  GIT_DEPLOY_REPO      Git repository URL
  GIT_DEPLOY_BRANCH    Git branch for deployment
        `);
        process.exit(0);
    }
    
    deployer.deploy().catch(console.error);
}

module.exports = DeploymentManager;