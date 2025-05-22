#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const https = require('https');

class SEOOptimizer {
    constructor() {
        this.rootDir = path.join(__dirname, '..');
        this.staticDir = path.join(this.rootDir, 'static');
        this.contentDir = path.join(this.rootDir, 'content');
        this.siteData = this.loadSiteData();
        this.seoMetrics = {};
    }

    loadSiteData() {
        const dataFile = path.join(this.contentDir, 'site-data.json');
        if (fs.existsSync(dataFile)) {
            return JSON.parse(fs.readFileSync(dataFile, 'utf8'));
        }
        return {};
    }

    async optimize() {
        console.log('Starting SEO optimization...');
        
        try {
            // Generate enhanced sitemaps
            await this.generateAdvancedSitemap();
            await this.generateRSSFeed();
            
            // Optimize meta tags and structured data
            await this.optimizeMetaTags();
            await this.addStructuredData();
            
            // Generate SEO reports
            await this.analyzeSEOMetrics();
            await this.generateSEOReport();
            
            // Submit to search engines
            await this.submitToSearchEngines();
            
            console.log('SEO optimization completed!');
            
        } catch (error) {
            console.error('SEO optimization failed:', error);
            throw error;
        }
    }

    async generateAdvancedSitemap() {
        const baseUrl = this.siteData.site?.url || 'https://autonomous-content.com';
        const pages = ['index', 'services', 'portfolio', 'contact', 'about'];
        const currentDate = new Date().toISOString().split('T')[0];
        
        // Main sitemap
        const sitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" 
        xmlns:image="http://www.google.com/schemas/sitemap-image/1.1"
        xmlns:news="http://www.google.com/schemas/sitemap-news/0.9">
${pages.map(page => {
    const url = page === 'index' ? baseUrl : `${baseUrl}/${page}.html`;
    const priority = page === 'index' ? '1.0' : page === 'services' ? '0.9' : '0.8';
    const changefreq = page === 'index' ? 'daily' : page === 'portfolio' ? 'weekly' : 'monthly';
    
    return `    <url>
        <loc>${url}</loc>
        <lastmod>${currentDate}</lastmod>
        <changefreq>${changefreq}</changefreq>
        <priority>${priority}</priority>
    </url>`;
}).join('\n')}
</urlset>`;

        fs.writeFileSync(path.join(this.staticDir, 'sitemap.xml'), sitemap);
        
        // Generate sitemap index for future expansion
        const sitemapIndex = `<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <sitemap>
        <loc>${baseUrl}/sitemap.xml</loc>
        <lastmod>${currentDate}</lastmod>
    </sitemap>
</sitemapindex>`;

        fs.writeFileSync(path.join(this.staticDir, 'sitemap-index.xml'), sitemapIndex);
        console.log('Generated advanced sitemap.xml and sitemap-index.xml');
    }

    async generateRSSFeed() {
        const baseUrl = this.siteData.site?.url || 'https://autonomous-content.com';
        const currentDate = new Date().toISOString();
        
        // Create RSS feed for portfolio updates
        const rss = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>
    <title>${this.siteData.site?.title} - Latest Work</title>
    <description>Latest content pieces created by our autonomous system</description>
    <link>${baseUrl}</link>
    <atom:link href="${baseUrl}/feed.xml" rel="self" type="application/rss+xml"/>
    <language>en-US</language>
    <lastBuildDate>${currentDate}</lastBuildDate>
    <generator>Autonomous Content System</generator>
    
${(this.siteData.portfolio || []).slice(0, 10).map(item => `    <item>
        <title>${item.title}</title>
        <description><![CDATA[${item.excerpt}]]></description>
        <link>${baseUrl}/portfolio.html#${item.id}</link>
        <guid>${baseUrl}/portfolio/${item.id}</guid>
        <pubDate>${new Date().toUTCString()}</pubDate>
        <category>${item.category}</category>
    </item>`).join('\n')}
</channel>
</rss>`;

        fs.writeFileSync(path.join(this.staticDir, 'feed.xml'), rss);
        console.log('Generated RSS feed');
    }

    async optimizeMetaTags() {
        const htmlFiles = fs.readdirSync(this.staticDir)
            .filter(file => file.endsWith('.html'));
        
        htmlFiles.forEach(file => {
            const filePath = path.join(this.staticDir, file);
            let html = fs.readFileSync(filePath, 'utf8');
            
            // Add Open Graph and Twitter Card meta tags if missing
            if (!html.includes('og:image')) {
                const ogImage = `${this.siteData.site?.url}/assets/images/og-image.png`;
                html = html.replace(
                    /<meta property="og:type" content="website">/,
                    `<meta property="og:type" content="website">
    <meta property="og:image" content="${ogImage}">
    <meta property="og:image:width" content="1200">
    <meta property="og:image:height" content="630">
    <meta name="twitter:image" content="${ogImage}">`
                );
            }
            
            // Add canonical URL
            if (!html.includes('rel="canonical"')) {
                const pageName = path.basename(file, '.html');
                const canonicalUrl = pageName === 'index' 
                    ? this.siteData.site?.url 
                    : `${this.siteData.site?.url}/${file}`;
                
                html = html.replace(
                    /<link rel="canonical"[^>]*>/,
                    `<link rel="canonical" href="${canonicalUrl}">`
                );
            }
            
            fs.writeFileSync(filePath, html);
        });
        
        console.log('Optimized meta tags for all HTML files');
    }

    async addStructuredData() {
        const organizationSchema = {
            "@context": "https://schema.org",
            "@type": "Organization",
            "name": this.siteData.site?.title,
            "description": this.siteData.site?.description,
            "url": this.siteData.site?.url,
            "logo": `${this.siteData.site?.url}/assets/images/logo.png`,
            "contactPoint": {
                "@type": "ContactPoint",
                "telephone": this.siteData.contact?.phone,
                "contactType": "customer service",
                "email": this.siteData.contact?.email
            },
            "sameAs": [
                // Add social media URLs when available
            ],
            "founder": {
                "@type": "Person",
                "name": "Autonomous AI System"
            },
            "foundingDate": "2024",
            "areaServed": "Worldwide",
            "serviceType": "Content Creation Services"
        };

        const serviceSchema = {
            "@context": "https://schema.org",
            "@type": "Service",
            "name": "Autonomous Content Creation",
            "description": "AI-powered content creation service operating 24/7",
            "provider": {
                "@type": "Organization",
                "name": this.siteData.site?.title
            },
            "serviceType": "Content Writing",
            "areaServed": "Worldwide",
            "hasOfferCatalog": {
                "@type": "OfferCatalog",
                "name": "Content Services",
                "itemListElement": (this.siteData.services || []).map((service, index) => ({
                    "@type": "Offer",
                    "itemOffered": {
                        "@type": "Service",
                        "name": service.name,
                        "description": service.description
                    },
                    "price": service.pricing,
                    "priceCurrency": "USD"
                }))
            }
        };

        // Add structured data to index page
        const indexPath = path.join(this.staticDir, 'index.html');
        if (fs.existsSync(indexPath)) {
            let html = fs.readFileSync(indexPath, 'utf8');
            
            const structuredData = `
    <script type="application/ld+json">
    ${JSON.stringify(organizationSchema, null, 2)}
    </script>
    <script type="application/ld+json">
    ${JSON.stringify(serviceSchema, null, 2)}
    </script>`;
            
            html = html.replace('</head>', `${structuredData}
</head>`);
            
            fs.writeFileSync(indexPath, html);
        }

        // Add review schema to portfolio items
        const portfolioPath = path.join(this.staticDir, 'portfolio.html');
        if (fs.existsSync(portfolioPath)) {
            let html = fs.readFileSync(portfolioPath, 'utf8');
            
            const portfolioSchema = {
                "@context": "https://schema.org",
                "@type": "CreativeWork",
                "name": "Content Portfolio",
                "description": "Examples of content created by autonomous AI system",
                "creator": {
                    "@type": "Organization",
                    "name": this.siteData.site?.title
                },
                "workExample": (this.siteData.portfolio || []).map(item => ({
                    "@type": "CreativeWork",
                    "name": item.title,
                    "description": item.excerpt,
                    "genre": item.category,
                    "wordCount": item.wordCount
                }))
            };
            
            const portfolioStructuredData = `
    <script type="application/ld+json">
    ${JSON.stringify(portfolioSchema, null, 2)}
    </script>`;
            
            html = html.replace('</head>', `${portfolioStructuredData}
</head>`);
            
            fs.writeFileSync(portfolioPath, html);
        }

        console.log('Added structured data to HTML files');
    }

    async analyzeSEOMetrics() {
        const htmlFiles = fs.readdirSync(this.staticDir)
            .filter(file => file.endsWith('.html'));
        
        htmlFiles.forEach(file => {
            const filePath = path.join(this.staticDir, file);
            const html = fs.readFileSync(filePath, 'utf8');
            const pageName = path.basename(file, '.html');
            
            this.seoMetrics[pageName] = {
                fileSize: fs.statSync(filePath).size,
                hasMetaDescription: html.includes('<meta name="description"'),
                hasMetaKeywords: html.includes('<meta name="keywords"'),
                hasOpenGraph: html.includes('property="og:'),
                hasTwitterCard: html.includes('name="twitter:'),
                hasStructuredData: html.includes('application/ld+json'),
                hasCanonical: html.includes('rel="canonical"'),
                headingCount: {
                    h1: (html.match(/<h1[^>]*>/g) || []).length,
                    h2: (html.match(/<h2[^>]*>/g) || []).length,
                    h3: (html.match(/<h3[^>]*>/g) || []).length
                },
                imageCount: (html.match(/<img[^>]*>/g) || []).length,
                linkCount: (html.match(/<a[^>]*href[^>]*>/g) || []).length,
                wordCount: this.countWords(html),
                loadTime: this.estimateLoadTime(fs.statSync(filePath).size)
            };
        });
        
        console.log('Analyzed SEO metrics for all pages');
    }

    countWords(html) {
        // Remove HTML tags and count words
        const text = html.replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim();
        return text ? text.split(' ').length : 0;
    }

    estimateLoadTime(fileSizeBytes) {
        // Estimate load time based on file size (rough calculation)
        const avgConnectionSpeed = 50000; // 50KB/s average
        return Math.round((fileSizeBytes / avgConnectionSpeed) * 1000); // milliseconds
    }

    async generateSEOReport() {
        const report = {
            generatedAt: new Date().toISOString(),
            summary: {
                totalPages: Object.keys(this.seoMetrics).length,
                averageLoadTime: Math.round(
                    Object.values(this.seoMetrics)
                        .reduce((sum, page) => sum + page.loadTime, 0) / 
                    Object.keys(this.seoMetrics).length
                ),
                seoScore: this.calculateOverallSEOScore()
            },
            pageMetrics: this.seoMetrics,
            recommendations: this.generateRecommendations()
        };
        
        const reportPath = path.join(this.staticDir, 'seo-report.json');
        fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
        
        // Generate human-readable report
        const readableReport = this.generateReadableReport(report);
        fs.writeFileSync(path.join(this.staticDir, 'seo-report.html'), readableReport);
        
        console.log('Generated SEO reports (JSON and HTML)');
        console.log(`Overall SEO Score: ${report.summary.seoScore}/100`);
    }

    calculateOverallSEOScore() {
        const pages = Object.values(this.seoMetrics);
        if (pages.length === 0) return 0;
        
        const totalScore = pages.reduce((sum, page) => {
            let pageScore = 0;
            
            // Meta tags (20 points)
            if (page.hasMetaDescription) pageScore += 10;
            if (page.hasOpenGraph) pageScore += 10;
            
            // Structured data (20 points)
            if (page.hasStructuredData) pageScore += 20;
            
            // Technical SEO (30 points)
            if (page.hasCanonical) pageScore += 10;
            if (page.loadTime < 3000) pageScore += 20; // Fast loading
            
            // Content structure (30 points)
            if (page.headingCount.h1 === 1) pageScore += 10; // Single H1
            if (page.headingCount.h2 > 0) pageScore += 10; // Has H2s
            if (page.wordCount > 300) pageScore += 10; // Sufficient content
            
            return sum + pageScore;
        }, 0);
        
        return Math.round(totalScore / pages.length);
    }

    generateRecommendations() {
        const recommendations = [];
        
        Object.entries(this.seoMetrics).forEach(([pageName, metrics]) => {
            if (!metrics.hasMetaDescription) {
                recommendations.push({
                    page: pageName,
                    priority: 'high',
                    issue: 'Missing meta description',
                    solution: 'Add a compelling meta description (150-160 characters)'
                });
            }
            
            if (metrics.loadTime > 3000) {
                recommendations.push({
                    page: pageName,
                    priority: 'medium',
                    issue: 'Slow page load time',
                    solution: 'Optimize images and minify CSS/JS'
                });
            }
            
            if (metrics.headingCount.h1 !== 1) {
                recommendations.push({
                    page: pageName,
                    priority: 'medium',
                    issue: `Incorrect H1 count: ${metrics.headingCount.h1}`,
                    solution: 'Use exactly one H1 tag per page'
                });
            }
            
            if (!metrics.hasStructuredData) {
                recommendations.push({
                    page: pageName,
                    priority: 'low',
                    issue: 'Missing structured data',
                    solution: 'Add Schema.org structured data for better search understanding'
                });
            }
        });
        
        return recommendations;
    }

    generateReadableReport(report) {
        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SEO Analysis Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; line-height: 1.6; }
        .header { background: #2563eb; color: white; padding: 20px; border-radius: 8px; }
        .score { font-size: 2em; font-weight: bold; }
        .metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin: 20px 0; }
        .metric-card { background: #f8fafc; padding: 20px; border-radius: 8px; border-left: 4px solid #2563eb; }
        .recommendations { margin-top: 30px; }
        .recommendation { background: #fff3cd; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #ffc107; }
        .high { border-left-color: #dc3545; background: #f8d7da; }
        .medium { border-left-color: #fd7e14; background: #fff3cd; }
        .low { border-left-color: #28a745; background: #d4edda; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background: #f8fafc; }
    </style>
</head>
<body>
    <div class="header">
        <h1>SEO Analysis Report</h1>
        <div class="score">Score: ${report.summary.seoScore}/100</div>
        <p>Generated: ${new Date(report.generatedAt).toLocaleString()}</p>
    </div>
    
    <div class="metrics">
        <div class="metric-card">
            <h3>Total Pages</h3>
            <div style="font-size: 2em; color: #2563eb;">${report.summary.totalPages}</div>
        </div>
        <div class="metric-card">
            <h3>Average Load Time</h3>
            <div style="font-size: 2em; color: #2563eb;">${report.summary.averageLoadTime}ms</div>
        </div>
    </div>
    
    <h2>Page Metrics</h2>
    <table>
        <thead>
            <tr>
                <th>Page</th>
                <th>File Size</th>
                <th>Load Time</th>
                <th>Word Count</th>
                <th>H1 Count</th>
                <th>Meta Description</th>
                <th>Structured Data</th>
            </tr>
        </thead>
        <tbody>
            ${Object.entries(report.pageMetrics).map(([page, metrics]) => `
            <tr>
                <td>${page}</td>
                <td>${Math.round(metrics.fileSize / 1024)}KB</td>
                <td>${metrics.loadTime}ms</td>
                <td>${metrics.wordCount}</td>
                <td>${metrics.headingCount.h1}</td>
                <td>${metrics.hasMetaDescription ? '✅' : '❌'}</td>
                <td>${metrics.hasStructuredData ? '✅' : '❌'}</td>
            </tr>
            `).join('')}
        </tbody>
    </table>
    
    <div class="recommendations">
        <h2>Recommendations</h2>
        ${report.recommendations.map(rec => `
        <div class="recommendation ${rec.priority}">
            <strong>${rec.page}</strong> - ${rec.issue}<br>
            <em>Solution: ${rec.solution}</em>
        </div>
        `).join('')}
    </div>
</body>
</html>`;
    }

    async submitToSearchEngines() {
        const siteUrl = this.siteData.site?.url;
        if (!siteUrl) {
            console.log('No site URL configured, skipping search engine submission');
            return;
        }
        
        const sitemapUrl = `${siteUrl}/sitemap.xml`;
        
        // Google Search Console (requires authentication)
        console.log(`Submit sitemap to Google: https://search.google.com/search-console`);
        console.log(`Sitemap URL: ${sitemapUrl}`);
        
        // Bing Webmaster Tools
        console.log(`Submit sitemap to Bing: https://www.bing.com/webmasters`);
        
        // Generate robots.txt if not exists
        const robotsPath = path.join(this.staticDir, 'robots.txt');
        if (!fs.existsSync(robotsPath)) {
            const robots = `User-agent: *
Allow: /

Sitemap: ${sitemapUrl}

# Block access to admin and private areas
Disallow: /admin/
Disallow: /private/
Disallow: /*.json$
Disallow: /seo-report.html`;
            
            fs.writeFileSync(robotsPath, robots);
            console.log('Generated robots.txt');
        }
    }
}

if (require.main === module) {
    const optimizer = new SEOOptimizer();
    optimizer.optimize().catch(console.error);
}

module.exports = SEOOptimizer;