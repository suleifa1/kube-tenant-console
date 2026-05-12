package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var dnsLabel = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

var deniedResources = map[string]bool{
	"secrets":                    true,
	"nodes":                      true,
	"persistentvolumes":          true,
	"clusterroles":               true,
	"clusterrolebindings":        true,
	"roles":                      true,
	"rolebindings":               true,
	"pods/exec":                  true,
	"pods/attach":                true,
	"pods/portforward":           true,
	"certificatesigningrequests": true,
}

func IsAssignmentExists(state *State, input Assignment) bool {
	for _, existing := range state.Assignments {
		if existing.TenantID == input.TenantID &&
			existing.NamespaceID == input.NamespaceID &&
			existing.RoleID == input.RoleID &&
			existing.SubjectKind == input.SubjectKind &&
			existing.SubjectName == input.SubjectName {
			return true
		}
	}
	return false
}

func ValidateSubjectKind(kind string) bool {
	switch kind {
	case "User", "Group", "ServiceAccount":
		return true
	default:
		return false
	}
}

func IsServiceAccountExists(s *State, namespaceId string, name string) bool {
	for _, sa := range s.ServiceAccounts {
		if sa.Name == name && sa.NamespaceID == namespaceId {
			return true
		}
	}
	return false
}

func ValidateName(kind, name string) error {
	if name == "" {
		return fmt.Errorf("%s name is required", kind)
	}
	if len(name) > 63 || !dnsLabel.MatchString(name) {
		return fmt.Errorf("%s name must be a DNS label", kind)
	}
	return nil
}

func ValidateRole(role RoleTemplate, cfg Config) error {
	if err := ValidateName("role", role.Name); err != nil {
		return err
	}
	if role.Scope == "" {
		role.Scope = "namespace"
	}
	if role.Scope != "namespace" && role.Scope != "cluster" {
		return fmt.Errorf("role scope must be namespace or cluster")
	}
	if role.Scope == "cluster" && !cfg.AllowClusterScope {
		return fmt.Errorf("cluster-scoped roles are disabled")
	}
	if len(role.Rules) == 0 {
		return fmt.Errorf("role must include at least one rule")
	}
	for _, rule := range role.Rules {
		if len(rule.Resources) == 0 || len(rule.Verbs) == 0 {
			return fmt.Errorf("each role rule needs resources and verbs")
		}
		for _, verb := range rule.Verbs {
			if strings.EqualFold(verb, "impersonate") {
				return fmt.Errorf("verb impersonate is denied")
			}
		}
		for _, resource := range rule.Resources {
			if deniedResources[strings.ToLower(resource)] {
				return fmt.Errorf("resource %q is denied by guardrails", resource)
			}
		}
	}
	return nil
}
