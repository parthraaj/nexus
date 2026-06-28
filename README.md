# Nexus

Platform for managing and operating OCP/Kubernetes clusters, node-level infrastructure, and IBM SVC storage.

Built on [kagent](https://kagent.dev/) (CNCF Sandbox) with custom MCP servers across 3 planes.

### Planes

**1. K8s:** Uses kagent built-in tools via `kagent-tool-server` for Kubernetes and OpenShift resource operations.

**2. Node**
- `node-daemon` - Privileged DaemonSet executing allowlisted host-level commands (iSCSI, multipath, NVMe) via nsenter
- `node-mcp-server` - MCP Streamable HTTP server routing node tool calls to node-daemon by pod IP

**3. Storage:** `storage-mcp-server` - MCP Streamable HTTP server for IBM SVC REST API operations

### Images
- `docker.io/parthrajghatge/nexus-node-daemon:v0.1.0`
- `docker.io/parthrajghatge/nexus-node-mcp-server:v0.1.0`
- `docker.io/parthrajghatge/nexus-storage-mcp-server:v0.1.0`
