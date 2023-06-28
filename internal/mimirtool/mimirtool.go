package mimirtool

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	toolName      = "mimirtool"
	addressFlag   = "--address="
	tenantFlag    = "--id="
	namespaceFlag = "--namespaces="
)

func callTool(ctx context.Context, args ...string) (string, string, error) {
	log.FromContext(ctx).Info("Running CLI", "parameters", args)

	cmd := exec.Command(toolName, args...)

	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return stderr.String(), "", err
	}

	return "", stdout.String(), nil
}

func getDefaultListArgs(tenant, url string) []string {
	addressParameter := addressFlag + url
	return []string{addressParameter, tenantFlag + tenant}
}

func getDefaultSyncArgs(tenant, url, namespace string) []string {
	namespaceParameter := namespaceFlag + namespace
	addressParameter := addressFlag + url

	return []string{addressParameter, tenantFlag + tenant, namespaceParameter}
}

func ListRules(ctx context.Context, tenant, url string) (string, error) {
	args := []string{"rules", "list"}
	args = append(args, getDefaultListArgs(tenant, url)...)
	args = append(args, "--format=json")
	args = append(args, "--disable-color")

	var stdout string
	var stderr string
	var err error

	if stderr, stdout, err = callTool(ctx, args...); err != nil {
		return "", fmt.Errorf("failed to call mimirtool cli: %s - %s", err.Error(), stderr)
	}

	return stdout, nil
}

func SynchronizeRule(ctx context.Context, ruleName, ruleFile, tenant, url string) error {
	args := []string{"rules", "sync"}
	args = append(args, getDefaultSyncArgs(tenant, url, ruleName)...)
	args = append(args, ruleFile)

	var stdout string
	var stderr string
	var err error

	if stderr, stdout, err = callTool(ctx, args...); err != nil {
		return fmt.Errorf("failed to call mimirtool cli: %s - %s", err.Error(), stderr)
	}

	log.FromContext(ctx).Info("CLI returned", "message", stdout)

	return nil
}
