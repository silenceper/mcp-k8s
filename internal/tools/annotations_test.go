package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestToolAnnotations(t *testing.T) {
	cases := []struct {
		tool        mcp.Tool
		readOnly    bool
		destructive bool
		idempotent  bool
	}{
		{CreateGetAPIResourcesTool(), true, false, true},
		{CreateGetResourceTool(), true, false, true},
		{CreateListResourcesTool(), true, false, true},
		{CreateGetPodLogsTool(), true, false, true},
		{CreateListEventsTool(), true, false, true},
		{CreateCreateResourceTool(), false, false, false},
		{CreateUpdateResourceTool(), false, true, true},
		{CreateDeleteResourceTool(), false, true, true},
		{CreateListHelmReleasesTool(), true, false, true},
		{CreateGetHelmReleaseTool(), true, false, true},
		{CreateInstallHelmChartTool(), false, false, false},
		{CreateUpgradeHelmChartTool(), false, true, false},
		{CreateUninstallHelmChartTool(), false, true, true},
		{CreateListHelmRepositoriesTool(), true, false, true},
		{CreateAddHelmRepositoryTool(), false, false, true},
		{CreateRemoveHelmRepositoryTool(), false, true, true},
	}

	for _, c := range cases {
		t.Run(c.tool.Name, func(t *testing.T) {
			a := c.tool.Annotations
			if a.ReadOnlyHint == nil || *a.ReadOnlyHint != c.readOnly {
				t.Errorf("readOnlyHint = %v, want %v", a.ReadOnlyHint, c.readOnly)
			}
			if a.DestructiveHint == nil || *a.DestructiveHint != c.destructive {
				t.Errorf("destructiveHint = %v, want %v", a.DestructiveHint, c.destructive)
			}
			if a.IdempotentHint == nil || *a.IdempotentHint != c.idempotent {
				t.Errorf("idempotentHint = %v, want %v", a.IdempotentHint, c.idempotent)
			}
			if a.OpenWorldHint == nil || !*a.OpenWorldHint {
				t.Errorf("openWorldHint = %v, want true", a.OpenWorldHint)
			}
		})
	}
}
