# Decision System Quick Reference

## Quick Start

### Making a Decision

```go
// Create decision request
request := decision_making.DecisionRequest{
    Type:        entities.DecisionTypeOperational,
    Priority:    entities.PriorityHigh,
    Title:       "Scale API Servers",
    Description: "Decision to scale API servers due to increased load",
    Context: map[string]interface{}{
        "current_load": 85,
        "target_load":  60,
        "cost_impact":  1000,
    },
}

// Make decision
decision, err := decisionService.MakeDecision(ctx, request)

// Execute if approved
if decision.Status == entities.StatusApproved {
    result, err := decisionService.ExecuteDecision(ctx, decision.ID)
}
```

### Creating a Policy

```go
policy := &entities.Policy{
    Name:        "API Rate Limits",
    Category:    "operational",
    Description: "Enforce API rate limiting rules",
    Priority:    100,
    Rules: []entities.PolicyRule{
        {
            ID:        "rate-limit-1",
            Condition: "requests_per_minute > 1000",
            Action:    "throttle",
            Severity:  "medium",
            Parameters: map[string]interface{}{
                "max_rate": 1000,
            },
        },
    },
    Active: true,
}

err := decisionService.RegisterPolicy(ctx, policy)
```

### Adding Ethical Guidelines

```go
guideline := &entities.EthicalGuideline{
    Principle:   "Data Privacy",
    Description: "Protect user data and privacy",
    Examples: []string{
        "Encrypt sensitive data",
        "Obtain explicit consent",
        "Allow data deletion",
    },
    RedLines: []string{
        "Selling user data",
        "Unauthorized data sharing",
        "Privacy violations",
    },
    Weight: 0.95,
}

err := decisionService.RegisterEthicalGuideline(ctx, guideline)
```

## Decision Types Reference

| Type | Auto-Approval | Confidence Required | Typical Time |
|------|---------------|-------------------|--------------|
| Operational | Yes | 90% | < 5 min |
| Strategic | No | N/A | 24-48 hrs |
| Financial | No | N/A | 1-2 hrs |
| Content | Yes | 85% | < 30 min |
| Client | Yes | 85% | < 1 hr |
| Emergency | Yes | Any | < 1 min |
| Ethical | Yes | 80% | < 30 min |
| Compliance | No | N/A | 2-4 hrs |

## API Endpoints

### Decision Management
```bash
POST   /api/v1/decisions              # Create decision
GET    /api/v1/decisions              # List decisions
GET    /api/v1/decisions/{id}         # Get decision
POST   /api/v1/decisions/{id}/execute # Execute decision
POST   /api/v1/decisions/{id}/override # Override decision
GET    /api/v1/decisions/{id}/quality # Assess quality
GET    /api/v1/decisions/{id}/logs    # Get audit logs
GET    /api/v1/decisions/metrics      # Get metrics
```

### Configuration
```bash
POST   /api/v1/policies               # Create policy
GET    /api/v1/policies               # List policies
POST   /api/v1/ethical-guidelines     # Add guideline
GET    /api/v1/ethical-guidelines     # List guidelines
```

### System
```bash
GET    /api/v1/system/health          # Health check
POST   /api/v1/system/emergency       # Emergency mode
GET    /api/v1/audit                  # Audit trail
```

## Decision Statuses

- `pending`: Awaiting analysis
- `analyzing`: Options being evaluated
- `approved`: Ready for execution
- `rejected`: Failed validation
- `overridden`: Manually changed
- `executed`: Completed
- `reverted`: Rolled back

## Policy Severity Levels

- `critical`: Immediate rejection, escalation required
- `high`: Rejection, notification sent
- `medium`: Warning, action required
- `low`: Information only

## Ethical Principles

1. **Do No Harm** (1.0) - Absolute priority
2. **Fairness** (0.9) - Equal treatment
3. **Transparency** (0.9) - Open operations
4. **Autonomy** (0.85) - Respect choice
5. **Beneficence** (0.8) - Create value
6. **Sustainability** (0.7) - Environmental care

## Common Patterns

### High-Risk Decision Pattern
```go
// For financial or strategic decisions
request := DecisionRequest{
    Type:     entities.DecisionTypeFinancial,
    Priority: entities.PriorityCritical,
    Context: map[string]interface{}{
        "amount": 50000,
        "risk_assessment_required": true,
        "approval_chain": []string{"treasury", "board"},
    },
}
```

### Emergency Response Pattern
```go
// For critical situations
if systemHealth < 0.5 {
    request := DecisionRequest{
        Type:     entities.DecisionTypeEmergency,
        Priority: entities.PriorityCritical,
        Title:    "Emergency Shutdown",
        Context: map[string]interface{}{
            "trigger": "system_failure",
            "severity": "critical",
        },
    }
}
```

### Batch Decision Pattern
```go
// For multiple related decisions
decisions := []DecisionRequest{
    {Type: entities.DecisionTypeContent, Title: "Publish Article 1"},
    {Type: entities.DecisionTypeContent, Title: "Publish Article 2"},
    {Type: entities.DecisionTypeContent, Title: "Publish Article 3"},
}

// Check for conflicts
conflicts := conflictResolver.DetectConflicts(ctx, decisions)
```

## Error Handling

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `policy violation` | Rule breach | Adjust request or policy |
| `ethical concern` | Red line hit | Change approach |
| `low confidence` | Weak options | Provide more context |
| `system unhealthy` | Low health score | Wait or emergency mode |
| `conflict detected` | Resource clash | Resolve or sequence |

### Error Response Format
```json
{
  "error": "policy_violation",
  "message": "Decision violates budget limit policy",
  "violations": [
    {
      "policy_id": "pol-001",
      "rule_id": "budget-max",
      "severity": "high"
    }
  ],
  "suggestions": [
    "Reduce budget to under $10,000",
    "Request override with justification"
  ]
}
```

## Performance Tips

1. **Cache Policies**: Policies are cached for 5 minutes
2. **Batch Operations**: Group related decisions
3. **Provide Context**: More context = better decisions
4. **Use Templates**: Leverage successful past decisions
5. **Monitor Metrics**: Track quality scores and improve

## Debugging

### Enable Debug Logging
```go
log.SetLevel("decision_making", "DEBUG")
```

### Common Debug Points
- Option generation
- Scoring calculations
- Policy evaluation
- Ethical validation
- Conflict detection

### Audit Trail Query
```sql
SELECT * FROM decision_logs 
WHERE decision_id = ? 
ORDER BY timestamp ASC;
```

## Best Practices

### DO
- ✅ Provide comprehensive context
- ✅ Define clear success criteria
- ✅ Monitor execution results
- ✅ Learn from failures
- ✅ Document overrides

### DON'T
- ❌ Skip ethical validation
- ❌ Ignore policy violations
- ❌ Execute without approval
- ❌ Bypass audit logging
- ❌ Ignore conflicts

## Emergency Procedures

### System Failure
```bash
POST /api/v1/system/emergency
{
  "reason": "Database connection lost"
}
```

### Manual Override
```bash
POST /api/v1/decisions/{id}/override
{
  "reason": "Market conditions require immediate action",
  "authorized_by": "admin@company.com"
}
```

### Rollback Decision
```go
err := decisionService.RevertDecision(ctx, decisionID, "Unexpected side effects")
```

## Metrics to Monitor

- **Decision Volume**: Decisions per hour/day
- **Confidence Distribution**: Average confidence scores
- **Override Rate**: Manual intervention frequency
- **Policy Compliance**: Violation rates by policy
- **Execution Success**: Success/failure rates
- **Response Time**: Decision-making speed

## Integration Examples

### With Content Pipeline
```go
decision := decisionService.MakeDecision(ctx, DecisionRequest{
    Type: entities.DecisionTypeContent,
    Title: "Publish Blog Post",
    Context: map[string]interface{}{
        "content_id": contentID,
        "quality_score": 0.92,
    },
})
```

### With Payment System
```go
decision := decisionService.MakeDecision(ctx, DecisionRequest{
    Type: entities.DecisionTypeFinancial,
    Title: "Process Payment",
    Context: map[string]interface{}{
        "amount": 5000,
        "currency": "USD",
        "client_id": clientID,
    },
})
```

### With Client Management
```go
decision := decisionService.MakeDecision(ctx, DecisionRequest{
    Type: entities.DecisionTypeClient,
    Title: "Accept New Project",
    Context: map[string]interface{}{
        "client_tier": "enterprise",
        "project_value": 50000,
        "capacity_available": true,
    },
})
```

## Troubleshooting Checklist

1. ☐ Check system health status
2. ☐ Verify policy compliance
3. ☐ Review ethical alignment
4. ☐ Examine conflict reports
5. ☐ Analyze confidence scores
6. ☐ Inspect audit logs
7. ☐ Review error messages
8. ☐ Check resource availability
9. ☐ Verify permissions
10. ☐ Test fallback plans