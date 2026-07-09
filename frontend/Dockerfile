# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Install corporate CA cert for TLS inspection proxy (Zscaler)
COPY zscaler-root-ca.crt /usr/local/share/ca-certificates/zscaler-root-ca.crt
RUN apk --no-cache add ca-certificates && update-ca-certificates
ENV NODE_EXTRA_CA_CERTS=/usr/local/share/ca-certificates/zscaler-root-ca.crt

COPY package.json package-lock.json* ./
RUN npm install

COPY . .
RUN npm run build

# Runtime stage - serve with nginx
FROM nginx:alpine

# Copy built assets
COPY --from=builder /app/dist /usr/share/nginx/html

# Custom nginx config template (uses envsubst for API_UPSTREAM)
COPY nginx.conf /etc/nginx/templates/default.conf.template

# Default upstream (matches service names in docker-compose and K8s)
ENV API_UPSTREAM="delivery-pulse-server:8090"

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
