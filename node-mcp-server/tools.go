package main

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerTools(s *server.MCPServer) {

	// ── READ: iSCSI sessions ──────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("node_get_iscsi_sessions",
			mcp.WithDescription("List active iSCSI sessions on a specific OCP/K8s node. Returns target IQN, portal IP, session state."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Full node name, e.g. ramen-ocp-3-1.xiv.ibm.com")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError("node parameter is required"), nil
			}
			out, err := callDaemon(ctx, node, "iscsi_sessions", nil)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)

	// ── READ: multipath topology ──────────────────────────────────────
	s.AddTool(
		mcp.NewTool("node_get_multipath",
			mcp.WithDescription("Show multipath device topology on a specific node. Returns map name, WWID, paths and their state (active/ghost/faulty)."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Full node name")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError("node parameter is required"), nil
			}
			out, err := callDaemon(ctx, node, "multipath_show", nil)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)

	// ── READ: block devices ───────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("node_get_lsblk",
			mcp.WithDescription("List all block devices on a specific node including filesystem type and mount points."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Full node name")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError("node parameter is required"), nil
			}
			out, err := callDaemon(ctx, node, "lsblk", nil)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)

	// ── WRITE/CREATE: iSCSI login ─────────────────────────────────────
	s.AddTool(
		mcp.NewTool("node_iscsi_login",
			mcp.WithDescription("Login to an iSCSI target on a specific node. MUTATING — requires approval. Creates a new iSCSI session."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Full node name")),
			mcp.WithString("target", mcp.Required(), mcp.Description("iSCSI target IQN, e.g. iqn.2001-04.com.ibm:storage.9.71.253.36")),
			mcp.WithString("portal", mcp.Required(), mcp.Description("Portal IP and port, e.g. 10.0.0.1:3260")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError("node parameter is required"), nil
			}
			target, err := req.RequireString("target")
			if err != nil {
				return mcp.NewToolResultError("target parameter is required"), nil
			}
			portal, err := req.RequireString("portal")
			if err != nil {
				return mcp.NewToolResultError("portal parameter is required"), nil
			}
			out, err := callDaemon(ctx, node, "iscsi_login", map[string]string{
				"target": target,
				"portal": portal,
			})
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)

	// ── WRITE/DELETE: iSCSI logout ────────────────────────────────────
	s.AddTool(
		mcp.NewTool("node_iscsi_logout",
			mcp.WithDescription("Logout from an iSCSI target on a specific node. MUTATING — requires approval. Removes the iSCSI session."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Full node name")),
			mcp.WithString("target", mcp.Required(), mcp.Description("iSCSI target IQN")),
			mcp.WithString("portal", mcp.Required(), mcp.Description("Portal IP and port, e.g. 10.0.0.1:3260")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError("node parameter is required"), nil
			}
			target, err := req.RequireString("target")
			if err != nil {
				return mcp.NewToolResultError("target parameter is required"), nil
			}
			portal, err := req.RequireString("portal")
			if err != nil {
				return mcp.NewToolResultError("portal parameter is required"), nil
			}
			out, err := callDaemon(ctx, node, "iscsi_logout", map[string]string{
				"target": target,
				"portal": portal,
			})
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)

	// ── WRITE/UPDATE: multipathd reconfigure ──────────────────────────
	s.AddTool(
		mcp.NewTool("node_multipath_reconfigure",
			mcp.WithDescription("Reconfigure multipathd on a specific node with current /etc/multipath.conf. MUTATING — requires approval."),
			mcp.WithString("node", mcp.Required(), mcp.Description("Full node name")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			node, err := req.RequireString("node")
			if err != nil {
				return mcp.NewToolResultError("node parameter is required"), nil
			}
			out, err := callDaemon(ctx, node, "multipath_reconfigure", nil)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return mcp.NewToolResultText(out), nil
		},
	)
}
