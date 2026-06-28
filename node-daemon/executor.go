package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

// ExecRequest is sent by node-mcp-server to node-daemon
type ExecRequest struct {
	Command string            `json:"command"`
	Params  map[string]string `json:"params"`
}

// ExecResponse is returned by node-daemon to node-mcp-server
type ExecResponse struct {
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
	IsError bool   `json:"is_error"`
}

func execute(req ExecRequest) ExecResponse {
	// Reject anything not in the allowlist
	cmd, ok := Allowlist[req.Command]
	if !ok {
		return ExecResponse{
			Error:   fmt.Sprintf("command %q is not in the allowlist", req.Command),
			IsError: true,
		}
	}

	// Validate all required params
	for _, key := range cmd.ParamKeys {
		val, exists := req.Params[key]
		if !exists {
			return ExecResponse{
				Error:   fmt.Sprintf("missing required param %q for command %q", key, req.Command),
				IsError: true,
			}
		}
		if err := sanitize(key, val); err != nil {
			return ExecResponse{Error: err.Error(), IsError: true}
		}
	}

	// Copy base args then append command-specific params
	args := make([]string, len(cmd.Args))
	copy(args, cmd.Args)

	switch req.Command {
	case "iscsi_login":
		// iscsiadm -m node -T <target> -p <portal> --login
		args = append(args, req.Params["target"], "-p", req.Params["portal"], "--login")
	case "iscsi_logout":
		// iscsiadm -m node -T <target> -p <portal> --logout
		args = append(args, req.Params["target"], "-p", req.Params["portal"], "--logout")
	}

	c := exec.Command(cmd.Binary, args...)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	if err != nil {
		combined := stdout.String()
		if stderr.Len() > 0 {
			combined += "\nSTDERR: " + stderr.String()
		}
		return ExecResponse{
			Output:  combined,
			Error:   err.Error(),
			IsError: true,
		}
	}

	return ExecResponse{Output: stdout.String()}
}
