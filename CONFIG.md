# Configuration File

cortextool supports a kubeconfig-style configuration file for managing multiple Cortex clusters with ease. This allows you to define multiple clusters, authentication credentials, and contexts, and easily switch between them.

## Quick Start

1. **Create a config file** at `~/.cortextool/config`:

```yaml
current-context: production

contexts:
  - name: production
    context:
      cluster: prod-cortex
      user: prod-user

clusters:
  - name: prod-cortex
    cluster:
      address: https://cortex.prod.example.com
      tls-ca-path: /path/to/ca.crt
      tls-cert-path: /path/to/client.crt
      tls-key-path: /path/to/client.key

users:
  - name: prod-user
    user:
      id: tenant-123
      auth-token: your-jwt-token
```

2. **Use cortextool commands** without specifying connection details:

```bash
# The address, tenant ID, and TLS certs are loaded from the config file
cortextool rules list
cortextool alertmanager get
```

3. **Switch between contexts**:

```bash
cortextool config use-context staging
```

## Configuration Structure

The configuration file consists of four main sections:

### 1. Current Context

Specifies which context to use by default:

```yaml
current-context: production
```

### 2. Contexts

Contexts bind together a cluster and user credentials:

```yaml
contexts:
  - name: production
    context:
      cluster: prod-cortex    # references a cluster name
      user: prod-user         # references a user name
  - name: staging
    context:
      cluster: staging-cortex
      user: staging-user
```

### 3. Clusters

Clusters define connection details for Cortex instances:

```yaml
clusters:
  - name: prod-cortex
    cluster:
      address: https://cortex.prod.example.com
      tls-ca-path: /path/to/ca.crt          # Optional: TLS CA certificate
      tls-cert-path: /path/to/client.crt    # Optional: Client certificate for mTLS
      tls-key-path: /path/to/client.key     # Optional: Client certificate key
      use-legacy-routes: false               # Optional: Use /api/prom/ routes
      ruler-api-path: /api/v1/rules         # Optional: Custom ruler API path
```

### 4. Users

Users define authentication credentials:

```yaml
users:
  - name: prod-user
    user:
      id: tenant-123              # Tenant ID
      auth-token: jwt-token-here  # Bearer token for JWT auth
      # OR use basic auth:
      # user: username
      # key: password
```

## Configuration Precedence

Configuration values are resolved in the following order (highest priority first):

1. **Command-line flags**: `--address`, `--id`, `--tls-cert-path`, etc.
2. **Environment variables**: `CORTEX_ADDRESS`, `CORTEX_TENANT_ID`, `CORTEX_TLS_CLIENT_CERT`, etc.
3. **`--context` flag**: Override which context to use (instead of `current-context`)
4. **Config file**: Values from the current context
5. **Defaults**: Built-in defaults

### Examples

```bash
# Use config file defaults (current-context)
cortextool rules list

# Override address with flag (other values still from config)
cortextool rules list --address https://different-cortex.com

# Override with environment variable
export CORTEX_TENANT_ID=different-tenant
cortextool rules list

# Use a different context temporarily with --context flag
cortextool --context staging rules list

# Combine --context with individual flags
cortextool --context production --id different-tenant rules list

# Use a different config file
cortextool --config /path/to/custom-config rules list
```

### Using the `--context` Flag

The `--context` flag allows you to temporarily use a different context without changing the `current-context` in your config file. This is useful for:

- **Quickly switching environments**: `cortextool --context staging rules list`
- **Testing with different credentials**: `cortextool --context test-user alertmanager get`
- **Scripts that need specific contexts**: Always use a specific context regardless of current-context

**Example workflow:**

```bash
# Set up multiple contexts
cortextool config set-context prod --cluster prod-cluster --user prod-user
cortextool config set-context staging --cluster staging-cluster --user staging-user
cortextool config use-context prod  # Set prod as default

# Use prod (current-context)
cortextool rules list

# Temporarily use staging without changing current-context
cortextool --context staging rules list

# Still using prod by default
cortextool rules list
```

**Error handling:**

If you specify a context that doesn't exist, cortextool will fail with a clear error:

```bash
cortextool --context nonexistent rules list
# Error: context "nonexistent" not found in config file
```

## Config Management Commands

cortextool provides subcommands to manage your configuration file:

### View Configuration

Display the current configuration:

```bash
cortextool config view
```

### List Contexts

Show all available contexts:

```bash
cortextool config get-contexts
```

Output:
```
CURRENT   NAME
*         production
          staging
          development
```

### Show Current Context

Display which context is currently active:

```bash
cortextool config current-context
```

### Switch Context

Change the active context:

```bash
cortextool config use-context staging
```

### Manage Contexts

Create or update a context:

```bash
# Create a new context
cortextool config set-context my-context --cluster my-cluster --user my-user

# Update an existing context
cortextool config set-context production --cluster new-prod-cluster
```

### Manage Clusters

Create or update cluster configuration:

```bash
cortextool config set-cluster prod-cortex \
  --address https://cortex.prod.example.com \
  --tls-ca-path /path/to/ca.crt \
  --tls-cert-path /path/to/client.crt \
  --tls-key-path /path/to/client.key
```

### Manage Credentials

Create or update user credentials:

```bash
# JWT authentication
cortextool config set-credentials prod-user \
  --id tenant-123 \
  --auth-token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Basic authentication
cortextool config set-credentials staging-user \
  --id tenant-456 \
  --user admin \
  --key password123
```

### Delete Context

Remove a context:

```bash
cortextool config delete-context old-context
```

## Default Config Location

By default, cortextool looks for the config file at:

```
$HOME/.cortextool/config
```

You can override this with:

- The `--config` flag: `cortextool --config /custom/path rules list`
- Set a custom default when creating the config file

## Example Workflows

### Multi-Cluster Setup

```bash
# Set up production cluster
cortextool config set-cluster prod --address https://cortex.prod.example.com
cortextool config set-credentials prod-user --id prod-tenant --auth-token <token>
cortextool config set-context prod --cluster prod --user prod-user

# Set up staging cluster
cortextool config set-cluster staging --address https://cortex.staging.example.com
cortextool config set-credentials staging-user --id staging-tenant --auth-token <token>
cortextool config set-context staging --cluster staging --user staging-user

# Use production
cortextool config use-context prod
cortextool rules list

# Switch to staging
cortextool config use-context staging
cortextool rules list
```

### Local Development

```bash
# Set up local development environment
cortextool config set-cluster local --address http://localhost:9009
cortextool config set-credentials dev --id dev
cortextool config set-context dev --cluster local --user dev
cortextool config use-context dev

# Now work with local Cortex
cortextool rules load rules.yaml
```

### Using TLS with mTLS

```bash
cortextool config set-cluster secure-cluster \
  --address https://cortex.secure.example.com \
  --tls-ca-path ~/.cortextool/certs/ca.crt \
  --tls-cert-path ~/.cortextool/certs/client.crt \
  --tls-key-path ~/.cortextool/certs/client.key

cortextool config set-credentials secure-user --id tenant-secure --auth-token <token>
cortextool config set-context secure --cluster secure-cluster --user secure-user
cortextool config use-context secure
```

## Migration from Environment Variables

If you're currently using environment variables, you can continue to use them alongside the config file. They will override config file values when set.

To migrate, convert your environment variables to a config file:

```bash
# Before (environment variables)
export CORTEX_ADDRESS=https://cortex.example.com
export CORTEX_TENANT_ID=my-tenant
export CORTEX_AUTH_TOKEN=my-token
cortextool rules list

# After (config file)
cortextool config set-cluster my-cluster --address https://cortex.example.com
cortextool config set-credentials my-user --id my-tenant --auth-token my-token
cortextool config set-context default --cluster my-cluster --user my-user
cortextool config use-context default
cortextool rules list  # No environment variables needed!
```

## Troubleshooting

### Config file not found

If cortextool can't find your config file, ensure it exists at `~/.cortextool/config` or specify it with `--config`.

### Invalid current context

If you get an error about the current context:

```bash
# Check available contexts
cortextool config get-contexts

# Switch to a valid context
cortextool config use-context <valid-context-name>
```

### Missing required fields

If a command fails with "cortex address is required" or "tenant ID is required", ensure your context references valid cluster and user entries with all required fields set.

## See Also

- [cortextool.example.yaml](./cortextool.example.yaml) - Full example configuration file
- [README.md](./README.md) - Main documentation
