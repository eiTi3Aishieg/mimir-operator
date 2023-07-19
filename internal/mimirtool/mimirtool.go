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
	tokenFlag     = "--auth-token"
	userFlag      = "--user"
	keyFlag       = "--key"
)

type Authentication struct {
	Username string
	Key      string
	Token    string
}

func callTool(ctx context.Context, auth *Authentication, args ...string) (string, string, error) {
	log.FromContext(ctx).Info("Running CLI", "parameters", args)

	cmd := exec.Command(toolName, args...)

	// Do not append the auth args to the other args BEFORE we log them as to not leak anything sensitive into the logs
	if auth != nil {
		authArgs := getAuthArgs(auth)
		args = append(args, authArgs...)
	}

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

func getAuthArgs(auth *Authentication) []string {
	if auth.Token != "" { // Token has precedence over anything else
		return []string{tokenFlag, auth.Token}
	} else {
		return []string{userFlag, auth.Username, keyFlag, auth.Key}
	}
}

func ListRules(ctx context.Context, auth *Authentication, tenant, url string) (string, error) {
	args := []string{"rules", "list"}
	args = append(args, getDefaultListArgs(tenant, url)...)
	args = append(args, "--format=json")
	args = append(args, "--disable-color")

	var stdout string
	var stderr string
	var err error

	if stderr, stdout, err = callTool(ctx, auth, args...); err != nil {
		return "", fmt.Errorf("failed to call mimirtool cli: %s - %s", err.Error(), stderr)
	}

	return stdout, nil
}

func SynchronizeRules(ctx context.Context, auth *Authentication, ruleName, ruleFile, tenant, url string) error {
	args := []string{"rules", "sync"}
	args = append(args, getDefaultSyncArgs(tenant, url, ruleName)...)
	args = append(args, ruleFile)

	var stdout string
	var stderr string
	var err error

	if stderr, stdout, err = callTool(ctx, auth, args...); err != nil {
		return fmt.Errorf("failed to call mimirtool cli: %s - %s", err.Error(), stderr)
	}

	log.FromContext(ctx).Info("CLI returned", "message", stdout)

	return nil
}
