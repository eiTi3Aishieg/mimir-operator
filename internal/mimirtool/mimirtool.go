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
	amSubCmd      = "alertmanager"
)

type Authentication struct {
	Username string
	Key      string
	Token    string
}

func callTool(ctx context.Context, auth *Authentication, args ...string) (string, string, error) {
	log.FromContext(ctx).Info("Running CLI", "parameters", args)

	// Do not append the auth args to the other args BEFORE we log them as to not leak anything sensitive into the logs
	if auth != nil {
		authArgs := getAuthArgs(auth)
		args = append(args, authArgs...)
	}

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
	tenantParameter := tenantFlag + tenant
	return []string{addressParameter, tenantParameter}
}

func getDefaultSyncArgs(tenant, url, namespace string) []string {
	addressParameter := addressFlag + url
	tenantParameter := tenantFlag + tenant
	namespaceParameter := namespaceFlag + namespace

	return []string{addressParameter, tenantParameter, namespaceParameter}
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

func VerifyAlertManagerConfig(ctx context.Context, auth *Authentication, configFile string) error {
	args := []string{amSubCmd, "verify"}
	args = append(args, configFile)

	var stderr string
	var err error

	// verify return error in stderr
	if stderr, _, err = callTool(ctx, auth, args...); err != nil {
		return fmt.Errorf("failed to call mimirtool cli: %s - %s", err.Error(), stderr)
	}

	log.FromContext(ctx).Info("CLI returned", "message", stderr)

	if stderr != "" {
		return fmt.Errorf("verification of config failed: \n%s", stderr)
	}
	return nil
}

func LoadAlertManagerConfig(ctx context.Context, auth *Authentication, configFile string, tenant, url string) error {
	args := []string{amSubCmd, "load"}
	args = append(args, getDefaultListArgs(tenant, url)...)
	args = append(args, configFile)

	var stdout string
	var stderr string
	var err error

	if stderr, stdout, err = callTool(ctx, auth, args...); err != nil {
		return fmt.Errorf("failed to call mimirtool cli: %s - %s", err.Error(), stderr)
	}

	log.FromContext(ctx).Info("CLI returned", "message", stdout)

	return nil
}

func DeleteAlertManagerConfig(ctx context.Context, auth *Authentication, tenant, url string) error {
	args := []string{amSubCmd, "delete"}
	args = append(args, getDefaultListArgs(tenant, url)...)

	var stdout string
	var stderr string
	var err error

	if stderr, stdout, err = callTool(ctx, auth, args...); err != nil {
		return fmt.Errorf("failed to call mimirtool cli: %s - %s", err.Error(), stderr)
	}

	log.FromContext(ctx).Info("CLI returned", "message", stdout)

	return nil
}
