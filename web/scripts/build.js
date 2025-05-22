#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { marked } = require('marked');
const { minify } = require('html-minifier');

class StaticSiteGenerator {
    constructor() {
        this.rootDir = path.join(__dirname, '..');
        this.templatesDir = path.join(this.rootDir, 'templates');
        this.assetsDir = path.join(this.rootDir, 'assets');
        this.outputDir = path.join(this.rootDir, 'static');
        this.contentDir = path.join(this.rootDir, 'content');
        
        this.ensureDirectories();
    }

    ensureDirectories() {
        [this.outputDir, this.contentDir].forEach(dir => {
            if (!fs.existsSync(dir)) {
                fs.mkdirSync(dir, { recursive: true });
            }
        });
    }

    loadTemplate(name) {
        const templatePath = path.join(this.templatesDir, `${name}.html`);
        if (!fs.existsSync(templatePath)) {
            throw new Error(`Template ${name} not found`);
        }
        return fs.readFileSync(templatePath, 'utf8');
    }

    processMarkdown(content) {
        return marked(content);
    }

    renderTemplate(template, data) {
        let rendered = template;
        
        Object.keys(data).forEach(key => {
            const pattern = new RegExp(`{{\\s*${key}\\s*}}`, 'g');
            rendered = rendered.replace(pattern, data[key] || '');
        });

        rendered = rendered.replace(/{{#each\s+(\w+)}}([\s\S]*?){{\/each}}/g, (match, arrayName, content) => {
            const array = data[arrayName] || [];
            return array.map(item => {
                let itemContent = content;
                Object.keys(item).forEach(itemKey => {
                    const itemPattern = new RegExp(`{{\\s*${itemKey}\\s*}}`, 'g');
                    itemContent = itemContent.replace(itemPattern, item[itemKey] || '');
                });
                return itemContent;
            }).join('');
        });

        return rendered;
    }

    async generatePage(pageName, data) {
        try {
            // Load base template and page template
            const baseTemplate = this.loadTemplate('base');
            const pageTemplate = this.loadTemplate(pageName);
            
            // Render page template with data
            const pageContent = this.renderTemplate(pageTemplate, data);
            
            // Inject page content into base template
            const pageData = {
                ...data,
                content: pageContent,
                pageName: pageName,
                pageTitle: this.getPageTitle(pageName, data),
                pageDescription: this.getPageDescription(pageName, data)
            };
            
            const html = this.renderTemplate(baseTemplate, pageData);
            
            const minified = minify(html, {
                removeComments: true,
                removeCommentsFromCDATA: true,
                removeCDATaSectionsFromCDATA: true,
                collapseWhitespace: true,
                conservativeCollapse: true,
                preserveLineBreaks: false,
                removeEmptyAttributes: true,
                removeOptionalTags: true,
                removeRedundantAttributes: true
            });

            const outputPath = path.join(this.outputDir, `${pageName}.html`);
            fs.writeFileSync(outputPath, minified);
            
            console.log(`Generated: ${outputPath}`);
            return outputPath;
        } catch (error) {
            console.error(`Error generating ${pageName}:`, error.message);
            throw error;
        }
    }

    getPageTitle(pageName, data) {
        const titles = {
            index: 'Home',
            services: 'Our Services',
            portfolio: 'Portfolio',
            contact: 'Contact Us',
            about: 'About Us'
        };
        return titles[pageName] || pageName.charAt(0).toUpperCase() + pageName.slice(1);
    }

    getPageDescription(pageName, data) {
        const descriptions = {
            index: data.site.description,
            services: 'Professional autonomous content creation services available 24/7',
            portfolio: 'Examples of high-quality content created by our autonomous AI system',
            contact: 'Get a quote for your content creation needs from our autonomous system',
            about: 'Learn about our pioneering autonomous content creation technology'
        };
        return descriptions[pageName] || data.site.description;
    }

    copyAssets() {
        if (!fs.existsSync(this.assetsDir)) return;
        
        const copyRecursive = (src, dest) => {
            if (!fs.existsSync(dest)) {
                fs.mkdirSync(dest, { recursive: true });
            }
            
            const items = fs.readdirSync(src);
            items.forEach(item => {
                const srcPath = path.join(src, item);
                const destPath = path.join(dest, item);
                
                if (fs.statSync(srcPath).isDirectory()) {
                    copyRecursive(srcPath, destPath);
                } else {
                    fs.copyFileSync(srcPath, destPath);
                }
            });
        };
        
        copyRecursive(this.assetsDir, path.join(this.outputDir, 'assets'));
        console.log('Assets copied to static directory');
    }

    async generateSitemap(pages) {
        const baseUrl = process.env.SITE_URL || 'https://autonomous-content.com';
        const currentDate = new Date().toISOString().split('T')[0];
        
        const sitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${pages.map(page => `    <url>
        <loc>${baseUrl}/${page === 'index' ? '' : page + '.html'}</loc>
        <lastmod>${currentDate}</lastmod>
        <changefreq>daily</changefreq>
        <priority>${page === 'index' ? '1.0' : '0.8'}</priority>
    </url>`).join('\n')}
</urlset>`;

        fs.writeFileSync(path.join(this.outputDir, 'sitemap.xml'), sitemap);
        console.log('Generated sitemap.xml');
    }

    async generateRobotsTxt() {
        const baseUrl = process.env.SITE_URL || 'https://autonomous-content.com';
        const robots = `User-agent: *
Allow: /

Sitemap: ${baseUrl}/sitemap.xml`;

        fs.writeFileSync(path.join(this.outputDir, 'robots.txt'), robots);
        console.log('Generated robots.txt');
    }

    async build() {
        console.log('Building autonomous web presence...');
        
        try {
            this.copyAssets();
            
            const pages = ['index', 'services', 'portfolio', 'contact', 'about', 'onboarding'];
            const siteData = this.loadSiteData();
            
            for (const page of pages) {
                await this.generatePage(page, siteData);
            }
            
            await this.generateSitemap(pages);
            await this.generateRobotsTxt();
            
            console.log('Build completed successfully!');
        } catch (error) {
            console.error('Build failed:', error);
            process.exit(1);
        }
    }

    loadSiteData() {
        const defaultData = {
            site: {
                title: 'Autonomous Content Creation Service',
                description: 'AI-powered content creation with unmatched quality and efficiency',
                url: process.env.SITE_URL || 'https://autonomous-content.com',
                keywords: 'AI content creation, autonomous writing, content marketing, SEO content'
            },
            services: [
                {
                    name: 'Blog Posts & Articles',
                    description: 'High-quality, SEO-optimized blog posts and articles',
                    pricing: 'Starting at $50'
                },
                {
                    name: 'Marketing Copy',
                    description: 'Compelling marketing copy that converts',
                    pricing: 'Starting at $75'
                },
                {
                    name: 'Technical Documentation',
                    description: 'Clear, comprehensive technical documentation',
                    pricing: 'Starting at $100'
                }
            ],
            portfolio: [],
            testimonials: [],
            contact: {
                email: 'hello@autonomous-content.com',
                phone: '+1 (555) 123-4567'
            }
        };

        const dataFile = path.join(this.contentDir, 'site-data.json');
        if (fs.existsSync(dataFile)) {
            try {
                const fileData = JSON.parse(fs.readFileSync(dataFile, 'utf8'));
                return { ...defaultData, ...fileData };
            } catch (error) {
                console.warn('Error loading site data, using defaults:', error.message);
            }
        }
        
        return defaultData;
    }
}

if (require.main === module) {
    const generator = new StaticSiteGenerator();
    generator.build().catch(console.error);
}

module.exports = StaticSiteGenerator;