package content_creation

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Ceesaxp/autonomous-content-service/src/domain/entities"
)

// PromptTemplate defines a template for generating prompts
type PromptTemplate struct {
	Name        string
	Template    string
	ContentType entities.ContentType
}

// PromptData contains data for filling prompt templates
type PromptData struct {
	ClientName     string
	ProjectTitle   string
	ContentTitle   string
	ContentType    entities.ContentType
	ContentGoals   []string
	TargetAudience string
	BrandVoice     string
	Keywords       []string
	StyleGuide     map[string]interface{}
	DomainKnowledge map[string]interface{}
	AdditionalContext map[string]interface{}
}

// PromptTemplateManager manages prompt templates for different content types
type PromptTemplateManager struct {
	templates map[entities.ContentType]map[string]*template.Template
}

// NewPromptTemplateManager creates a new prompt template manager
func NewPromptTemplateManager() *PromptTemplateManager {
	manager := &PromptTemplateManager{
		templates: make(map[entities.ContentType]map[string]*template.Template),
	}
	
	// Register default templates
	manager.registerDefaultTemplates()
	
	return manager
}

// registerDefaultTemplates registers the default templates for different content types
func (m *PromptTemplateManager) registerDefaultTemplates() {
	// Blog post templates
	m.RegisterTemplate(entities.ContentTypeBlogPost, "research", blogPostResearchTemplate)
	m.RegisterTemplate(entities.ContentTypeBlogPost, "outline", blogPostOutlineTemplate)
	m.RegisterTemplate(entities.ContentTypeBlogPost, "draft", blogPostDraftTemplate)
	m.RegisterTemplate(entities.ContentTypeBlogPost, "edit", blogPostEditTemplate)
	m.RegisterTemplate(entities.ContentTypeBlogPost, "finalize", blogPostFinalizeTemplate)
	
	// Social post templates
	m.RegisterTemplate(entities.ContentTypeSocialPost, "research", socialPostResearchTemplate)
	m.RegisterTemplate(entities.ContentTypeSocialPost, "draft", socialPostDraftTemplate)
	m.RegisterTemplate(entities.ContentTypeSocialPost, "edit", socialPostEditTemplate)
	
	// Technical article templates
	m.RegisterTemplate(entities.ContentTypeTechnicalArticle, "research", technicalArticleResearchTemplate)
	m.RegisterTemplate(entities.ContentTypeTechnicalArticle, "outline", technicalArticleOutlineTemplate)
	m.RegisterTemplate(entities.ContentTypeTechnicalArticle, "draft", technicalArticleDraftTemplate)
	m.RegisterTemplate(entities.ContentTypeTechnicalArticle, "edit", technicalArticleEditTemplate)
	m.RegisterTemplate(entities.ContentTypeTechnicalArticle, "finalize", technicalArticleFinalizeTemplate)
	
	// Add other content types as needed...
}

// RegisterTemplate registers a new template
func (m *PromptTemplateManager) RegisterTemplate(contentType entities.ContentType, name, templateText string) error {
	// Initialize the map for this content type if it doesn't exist
	if _, exists := m.templates[contentType]; !exists {
		m.templates[contentType] = make(map[string]*template.Template)
	}
	
	// Parse the template
	tmpl, err := template.New(name).Parse(templateText)
	if err != nil {
		return err
	}
	
	// Store the template
	m.templates[contentType][name] = tmpl
	
	return nil
}

// GeneratePrompt generates a prompt using the specified template
func (m *PromptTemplateManager) GeneratePrompt(contentType entities.ContentType, templateName string, data PromptData) (string, error) {
	// Check if template exists
	contentTemplates, exists := m.templates[contentType]
	if !exists {
		return "", fmt.Errorf("no templates found for content type: %s", contentType)
	}
	
	tmpl, exists := contentTemplates[templateName]
	if !exists {
		return "", fmt.Errorf("template not found: %s for content type: %s", templateName, contentType)
	}
	
	// Execute the template
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	
	return buf.String(), nil
}

// Template constants - these would contain the actual template strings
const (
	blogPostResearchTemplate = `You are conducting research for a blog post titled "{{.ContentTitle}}" for {{.ClientName}}.
The target audience is {{.TargetAudience}}.
The main content goals are: {{range .ContentGoals}}
- {{.}}{{end}}

Based on this information, identify key topics, relevant facts, and statistics that should be included in the blog post.
Format your response as a structured research document with main topics, supporting points, and potential sources.
{{if .DomainKnowledge}}Additional context to consider:
{{range $key, $value := .DomainKnowledge}}
{{$key}}: {{$value}}{{end}}{{end}}`

	blogPostOutlineTemplate = `Create a detailed outline for a blog post titled "{{.ContentTitle}}" for {{.ClientName}}.
The target audience is {{.TargetAudience}} and the post should align with their {{.BrandVoice}} brand voice.
The content should incorporate these keywords: {{range .Keywords}}{{.}}, {{end}}

The outline should include:
1. Introduction with a compelling hook
2. Main sections with subpoints (at least 3-5 main sections)
3. Conclusion with call to action

Format the outline with clear hierarchical structure using headings and subheadings.`

	blogPostDraftTemplate = `Write a comprehensive blog post draft titled "{{.ContentTitle}}" for {{.ClientName}}.
Follow this outline:
{{.AdditionalContext.Outline}}

The content should be written in a {{.BrandVoice}} tone for {{.TargetAudience}}.
Naturally incorporate these keywords: {{range .Keywords}}{{.}}, {{end}}

Include relevant examples, data points, and actionable advice. The content should be engaging, informative, and aligned with these goals: {{range .ContentGoals}}
- {{.}}{{end}}`

	blogPostEditTemplate = `Edit the following blog post draft to improve readability, flow, accuracy, and engagement.
Title: {{.ContentTitle}}
Client: {{.ClientName}}
Audience: {{.TargetAudience}}
Brand Voice: {{.BrandVoice}}

Draft to edit:
{{.AdditionalContext.Draft}}

Focus on:
1. Improving sentence structure and flow between paragraphs
2. Enhancing clarity and readability
3. Ensuring consistent tone and style
4. Strengthening the introduction and conclusion
5. Verifying factual accuracy
6. Naturally incorporating these keywords: {{range .Keywords}}{{.}}, {{end}}`

	blogPostFinalizeTemplate = `Finalize the following blog post for publication.
Title: {{.ContentTitle}}
Client: {{.ClientName}}

Content:
{{.AdditionalContext.EditedDraft}}

Format the post for web publication with:
- Proper heading structure (H1, H2, H3)
- Short, scannable paragraphs
- Strategic use of bold text for emphasis
- SEO optimization for the target keywords: {{range .Keywords}}{{.}}, {{end}}
- Internal linking suggestions (placeholder URLs)
- Meta description suggestion (under 160 characters)
- Social sharing snippet (under 100 characters)`

	// Add templates for other content types...
	socialPostResearchTemplate = `Research engaging social media content related to "{{.ContentTitle}}" for {{.ClientName}}.
Identify trending topics, hashtags, and engagement patterns in the {{.AdditionalContext.Platform}} platform.
The target audience is {{.TargetAudience}}.
Identify viral content patterns and engagement triggers for this audience.`

	socialPostDraftTemplate = `Create a compelling social media post for {{.ClientName}} about "{{.ContentTitle}}".
The post should be written in a {{.BrandVoice}} tone for {{.TargetAudience}} and suitable for {{.AdditionalContext.Platform}}.
Maximum length: {{.AdditionalContext.MaxLength}} characters.
Include appropriate hashtags and call to action.`

	socialPostEditTemplate = `Refine the following social media post to maximize engagement while maintaining brand voice.
Platform: {{.AdditionalContext.Platform}}
Brand: {{.ClientName}}
Audience: {{.TargetAudience}}
Tone: {{.BrandVoice}}

Original post:
{{.AdditionalContext.Draft}}

Consider:
- Emotional impact and shareability
- Call-to-action effectiveness
- Character count optimization
- Hashtag relevance and quantity
- Visual description suggestions if applicable`

	technicalArticleResearchTemplate = `Conduct technical research for an article titled "{{.ContentTitle}}" for {{.ClientName}}.
The target audience consists of {{.TargetAudience}}.
The article should cover these technical concepts: {{range .AdditionalContext.TechnicalConcepts}}{{.}}, {{end}}

Identify key technical information, specifications, code examples, and authoritative sources.
Focus on accuracy, technical depth, and educational value for the target audience.`

	technicalArticleOutlineTemplate = `Create a detailed outline for a technical article titled "{{.ContentTitle}}" for {{.ClientName}}.
The audience consists of {{.TargetAudience}} with knowledge in {{.AdditionalContext.TechnicalDomain}}.

The outline should include:
1. Introduction with technical context and article purpose
2. Technical background/prerequisites
3. Main technical concepts (at least 3-5 major sections with subsections)
4. Practical implementation or examples
5. Technical considerations, limitations, or trade-offs
6. Conclusion with key technical takeaways

Include placeholders for code examples, diagrams, or technical illustrations where appropriate.`

	technicalArticleDraftTemplate = `Write a comprehensive technical article titled "{{.ContentTitle}}" for {{.ClientName}}.
Follow this technical outline:
{{.AdditionalContext.Outline}}

The article should:
- Maintain technical accuracy and precision
- Include clear technical explanations for {{.TargetAudience}}
- Incorporate code examples using proper syntax formatting
- Reference technical specifications and standards where appropriate
- Use proper technical terminology consistent with {{.AdditionalContext.TechnicalDomain}}

The content should be technically informative while maintaining readability.`

	technicalArticleEditTemplate = `Review and improve the following technical article draft for technical accuracy, clarity, and educational value.
Title: {{.ContentTitle}}
Client: {{.ClientName}}
Audience: {{.TargetAudience}}
Technical Domain: {{.AdditionalContext.TechnicalDomain}}

Draft to edit:
{{.AdditionalContext.Draft}}

Focus on:
1. Technical accuracy and precision
2. Clarity of technical explanations
3. Code example correctness and best practices
4. Technical terminology consistency
5. Logical flow of technical concepts
6. Appropriate level of technical detail for the audience`

	technicalArticleFinalizeTemplate = `Finalize the following technical article for publication.
Title: {{.ContentTitle}}
Client: {{.ClientName}}
Technical Domain: {{.AdditionalContext.TechnicalDomain}}

Content:
{{.AdditionalContext.EditedDraft}}

Prepare for technical publication with:
- Properly formatted technical headings and subheadings
- Correctly formatted code blocks with syntax highlighting hints
- Technical diagrams or illustration placeholders where needed
- Properly formatted technical references/citations
- Technical glossary for key terms (if appropriate)
- Table of contents for navigation
- SEO optimization for technical search terms`
)