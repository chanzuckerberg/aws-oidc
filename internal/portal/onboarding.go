package portal

import "net/http"

// onboardingParam marks a request as part of the post-create walkthrough. It travels as a query
// parameter on a GET and a hidden field on a POST, so the walkthrough needs no server-side state.
const onboardingParam = "onboarding"

// onboardingStep is one page of the walkthrough.
type onboardingStep struct {
	Nav   string
	Label string
	Path  string
}

// onboarding is the walkthrough state handed to a template. The zero value means the page is
// being edited on its own rather than as part of setup.
type onboarding struct {
	Active bool
	Steps  []onboardingStep
	// Index is the zero-based position of the page being rendered.
	Index int
	// Next is the path of the following step, empty on the last one.
	Next string
	// Done is where the walkthrough sends the browser after its last step.
	Done string
}

func (o onboarding) Number() int { return o.Index + 1 }
func (o onboarding) Total() int  { return len(o.Steps) }
func (o onboarding) Last() bool  { return o.Index == len(o.Steps)-1 }
func (o onboarding) NextURL() string {
	return o.Next + "?" + onboardingParam + "=1"
}

// onboardingSteps returns the walkthrough in order, leaving out the pages this portal does not
// offer. AWS access comes last because it is the one page that waits on Okta, which gives the
// lookup the earlier steps to finish in.
func (s *Server) onboardingSteps(agent string) []onboardingStep {
	prefix := "/agents/" + agent
	steps := make([]onboardingStep, 0, 4)
	if s.cfg.AgentRuntime {
		steps = append(steps, onboardingStep{Nav: "runtime", Label: "Runtime", Path: prefix + "/runtime"})
	}
	if s.cfg.AgentTailscale {
		steps = append(steps, onboardingStep{Nav: "tailscale", Label: "Tailscale", Path: prefix + "/tailscale"})
	}
	if s.cfg.Repositories != nil {
		steps = append(steps, onboardingStep{Nav: "repositories", Label: "Repositories", Path: prefix + "/repositories"})
	}
	steps = append(steps, onboardingStep{Nav: "aws", Label: "AWS access", Path: prefix + "/aws"})
	return steps
}

// onboardingStart is where a newly created agent sends its owner.
func (s *Server) onboardingStart(agent string) string {
	steps := s.onboardingSteps(agent)
	return steps[0].Path + "?" + onboardingParam + "=1"
}

// onboardingDone is where the walkthrough lands once its last step is saved. Connection is the
// payoff page, so an owner whose agent runs in the cluster ends on the command that reaches it.
func (s *Server) onboardingDone(agent string) string {
	if s.cfg.AgentRuntime {
		return "/agents/" + agent + "/connection"
	}
	return "/agents/" + agent
}

// onboardingFor returns walkthrough state for one page. It is inactive unless the request
// carries the marker and the page is a step this portal offers.
func (s *Server) onboardingFor(r *http.Request, agent, nav string) onboarding {
	if r.FormValue(onboardingParam) != "1" {
		return onboarding{}
	}
	steps := s.onboardingSteps(agent)
	for i, step := range steps {
		if step.Nav != nav {
			continue
		}
		o := onboarding{Active: true, Steps: steps, Index: i, Done: s.onboardingDone(agent)}
		if i+1 < len(steps) {
			o.Next = steps[i+1].Path
		}
		return o
	}
	return onboarding{}
}

// redirectAfterSave sends the browser on to the next walkthrough step, or back to the page it
// just saved when the page is being edited on its own.
func (s *Server) redirectAfterSave(w http.ResponseWriter, r *http.Request, agent, nav string) {
	o := s.onboardingFor(r, agent, nav)
	switch {
	case !o.Active:
		s.redirect(w, r, "/agents/"+agent+"/"+nav)
	case o.Next == "":
		s.redirect(w, r, o.Done)
	default:
		s.redirect(w, r, o.NextURL())
	}
}
