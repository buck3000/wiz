package license

import "fmt"

// Tier represents a wiz licensing tier.
type Tier int

const (
	TierFree       Tier = iota
	TierPro             // $29/mo — power users
	TierTeam            // $99/mo per seat — teams
	TierEnterprise      // custom pricing
)

func (t Tier) String() string {
	switch t {
	case TierFree:
		return "Free"
	case TierPro:
		return "Pro"
	case TierTeam:
		return "Team"
	case TierEnterprise:
		return "Enterprise"
	default:
		return fmt.Sprintf("Tier(%d)", int(t))
	}
}

// Limits defines feature gates for a tier.
type Limits struct {
	MaxContexts   int  // 0 = unlimited
	OrchestraDeps bool // depends_on in orchestra files
	CostTracking  bool
	AIReview      bool // wiz review (LLM-as-judge)
	TeamRegistry  bool // shared context registry
	CIMode        bool // wiz orchestra --ci --headless
	AuditLog      bool
}

// LimitsForTier returns the feature limits for the given tier.
func LimitsForTier(t Tier) Limits {
	switch t {
	case TierPro:
		return Limits{
			MaxContexts:   0, // unlimited
			OrchestraDeps: true,
			CostTracking:  true,
		}
	case TierTeam:
		return Limits{
			MaxContexts:   0,
			OrchestraDeps: true,
			CostTracking:  true,
			AIReview:      true,
			TeamRegistry:  true,
			CIMode:        true,
			AuditLog:      true,
		}
	case TierEnterprise:
		return Limits{
			MaxContexts:   0,
			OrchestraDeps: true,
			CostTracking:  true,
			AIReview:      true,
			TeamRegistry:  true,
			CIMode:        true,
			AuditLog:      true,
		}
	default: // Free
		return Limits{
			MaxContexts: 10,
		}
	}
}

// ContextLimitErr is returned when the context limit is reached.
type ContextLimitErr struct {
	Current int
	Max     int
	Tier    Tier
}

func (e *ContextLimitErr) Error() string {
	return fmt.Sprintf("context limit reached (%d/%d) on Wiz %s", e.Current, e.Max, e.Tier)
}

// CheckContextLimit returns a ContextLimitErr if creating another context
// would exceed the tier's limit. Returns nil if allowed.
func CheckContextLimit(tier Tier, currentCount int) error {
	limits := LimitsForTier(tier)
	if limits.MaxContexts > 0 && currentCount >= limits.MaxContexts {
		return &ContextLimitErr{
			Current: currentCount,
			Max:     limits.MaxContexts,
			Tier:    tier,
		}
	}
	return nil
}
