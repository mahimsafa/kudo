# Deploying a Node.js Application as a Docker Container

This guide walks you through containerizing a Node.js app and deploying it on a Kudo cluster using the Docker adapter.

## Prerequisites

- A running Kudo cluster (see [Getting Started](getting-started.md))
- Docker installed on all cluster nodes
- A Node.js application with a `Dockerfile`

## 1. Prepare Your Node.js App

Your project structure should look something like:

```
my-node-app/
├── package.json
├── package-lock.json
├── src/
│   └── index.js
└── Dockerfile
```

### Example `package.json`

```json
{
  "name": "my-node-app",
  "version": "1.0.0",
  "scripts": {
    "start": "node src/index.js"
  },
  "dependencies": {
    "express": "^4.18.0"
  }
}
```

### Example `src/index.js`

```javascript
const express = require("express");
const app = express();
const PORT = process.env.PORT || 3000;

app.get("/health", (req, res) => {
  res.json({ status: "ok" });
});

app.get("/", (req, res) => {
  res.json({ message: "Hello from Kudo!", hostname: require("os").hostname() });
});

app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});
```

## 2. Write the Dockerfile

```dockerfile
FROM node:20-alpine

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci --production

COPY src/ ./src/

ENV PORT=3000
EXPOSE 3000

USER node
CMD ["npm", "start"]
```

## 3. Build and Push the Image

Build and push to your container registry:

```bash
docker build -t registry.example.com/my-node-app:v1.0 .
docker push registry.example.com/my-node-app:v1.0
```

Or if you're using Docker Hub:

```bash
docker build -t yourusername/my-node-app:v1.0 .
docker push yourusername/my-node-app:v1.0
```

Make sure the image is accessible from all cluster nodes.

## 4. Create the Kudo Application Config

Create `my-node-app.yaml`:

```yaml
kind: Application
name: my-node-app
adapter: docker
replicas: 3
spec:
  image: registry.example.com/my-node-app:v1.0
  env:
    PORT: "3000"
    NODE_ENV: "production"
  ports:
    - 3000
routing:
  domain: app.example.com
  path: /
  algorithm: round-robin
  healthcheck:
    path: /health
    interval: 10s
    timeout: 3s
    unhealthy_threshold: 3
```

### Config Breakdown

| Field | Description |
|-------|-------------|
| `adapter: docker` | Uses the built-in Docker executor |
| `replicas: 3` | Runs 3 container instances spread across nodes |
| `spec.image` | The Docker image to pull and run |
| `spec.env` | Environment variables passed to the container |
| `spec.ports` | Ports to expose (Kudo assigns random host ports) |
| `routing.domain` | Domain name for the L7 proxy to route traffic |
| `routing.healthcheck.path` | Endpoint the proxy uses to verify the container is healthy |

## 5. Deploy

```bash
kudo apply -f my-node-app.yaml
```

## 6. Verify

Check that the app is running:

```bash
kudo status my-node-app
```

Expected output:

```
Application: my-node-app
Adapter:     docker
Replicas:    3/3 running

Instances:
ID        NODE      STATUS   ADDRESS
a1b2c3d4  node-1    running  127.0.0.1:49152
e5f6g7h8  node-2    running  127.0.0.1:49153
i9j0k1l2  node-3    running  127.0.0.1:49154
```

## 7. Access the App

Point your DNS for `app.example.com` to any cluster node's IP. The Kudo L7 proxy listens on port 80 and round-robins requests across all healthy instances.

```bash
curl -H "Host: app.example.com" http://<any-node-ip>/
# {"message":"Hello from Kudo!","hostname":"kudo-my-node-app-a1b2c3d4"}
```

Each request hits a different container, confirmed by the changing hostname.

## 8. Scale

Scale up or down at any time:

```bash
kudo scale my-node-app --replicas 5
```

The reconciler will schedule new containers on nodes with the fewest instances (spread strategy).

## 9. Update the Image

To deploy a new version, update the image tag in your config and re-apply:

```yaml
spec:
  image: registry.example.com/my-node-app:v1.1
```

```bash
kudo apply -f my-node-app.yaml
```

## Troubleshooting

**Image pull failures** — Ensure the image registry is accessible from all nodes and credentials are configured in Docker (`docker login`).

**Health check failures** — Verify your `/health` endpoint returns a 200 status and that the `PORT` environment variable matches `spec.ports`.

**Container crashes** — Check container logs via Docker directly:

```bash
docker logs kudo-my-node-app-<instance-id>
```

## Next Steps

- [Configuration Reference](configuration.md) — full list of app config options
- [Architecture](architecture.md) — how the scheduler and proxy work together
