# ssh-sync-server

ssh-sync-server is the companion server component to the [ssh-sync](https://github.com/therealpaulgg/ssh-sync) client application. It securely stores client-encrypted SSH keys and configurations, enabling seamless synchronization across multiple machines.

[![release](https://github.com/therealpaulgg/ssh-sync-server/actions/workflows/release.yml/badge.svg)](https://github.com/therealpaulgg/ssh-sync-server/actions/workflows/release.yml)

## Introduction

ssh-sync-server provides the backend infrastructure for the ssh-sync ecosystem. While users interact with the ssh-sync client to manage their SSH keys locally, this server component:

- Securely stores encrypted SSH keys and configurations
- Facilitates synchronization between multiple devices
- Manages authentication and machine registration
- Provides APIs for the ssh-sync client to communicate with

All data stored on the server is encrypted by the client before transmission, ensuring that even server administrators cannot access your private keys.

## Quick Start

The fastest way to get started with ssh-sync-server is using Docker:

```bash
# Pull the official Docker image
docker pull therealpaulgg/ssh-sync-server:latest
docker pull therealpaulgg/ssh-sync-db:latest

# Create a simple docker-compose.yml file and start the services
docker-compose up -d
```

## Self-Hosting Guide

### Prerequisites

- Docker and Docker Compose
- A domain name (for production setups)
- Basic knowledge of reverse proxies and SSL certificates (for production)

### Docker Setup

Here's a complete `docker-compose.yaml` example for a production environment:

```yaml
version: '3.3'
services:
    ssh-sync-server:
        restart: always
        environment:
          - PORT=3000
          - NO_DOTENV=1
          - DATABASE_USERNAME=sshsync
          - DATABASE_PASSWORD=${POSTGRES_PASSWORD}
          - DATABASE_NAME=sshsync
          - DATABASE_HOST=ssh-sync-db:5432
        logging:
          driver: json-file
          options:
            max-size: 10m
        ports:
          - '127.0.0.1:3000:3000'  # Only bind to localhost if behind reverse proxy
        image: therealpaulgg/ssh-sync-server:latest
        container_name: ssh-sync-server
    ssh-sync-db:
        image: therealpaulgg/ssh-sync-db:latest
        container_name: ssh-sync-db
        volumes:
          - /path/to/db-volume:/var/lib/postgresql/data
        environment:
          - POSTGRES_USER=sshsync
          - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
          - POSTGRES_DB=sshsync
        restart: always
```

Save this file and run:

```bash
export POSTGRES_PASSWORD=your_secure_password_here
docker-compose up -d
```

### Environment Variables

The server can be configured using the following environment variables:

| Variable | Description | Default |
| -------- | ----------- | ------- |
| PORT | The port the server will listen on | 3000 |
| NO_DOTENV | Set to "1" to disable loading from .env file | (unset) |
| DATABASE_USERNAME | PostgreSQL database username | N/A |
| DATABASE_PASSWORD | PostgreSQL database password | N/A |
| DATABASE_NAME | PostgreSQL database name | N/A |
| DATABASE_HOST | PostgreSQL host address | N/A |

### Setting Up with Nginx Reverse Proxy

For production environments, we recommend using a reverse proxy like Nginx with SSL certificates from Let's Encrypt.

Example Nginx configuration (must support websockets):

```nginx
server {
    listen [::]:443 ssl ipv6only=on; # managed by Certbot
    listen 443 ssl; # managed by Certbot
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem; # managed by Certbot
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem; # managed by Certbot
    include /etc/letsencrypt/options-ssl-nginx.conf; # managed by Certbot
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem; # managed by Certbot
    server_name your-domain.com;
    location / {
          proxy_pass http://127.0.0.1:3000;
          proxy_http_version 1.1;
          proxy_set_header Upgrade $http_upgrade;
          proxy_set_header Connection "Upgrade";
          proxy_set_header Host $host;
          proxy_set_header X-Forwarded-For $remote_addr;
          proxy_set_header X-Real-IP $remote_addr;
    }
}

server {
    if ($host = your-domain.com) {
        return 301 https://$host$request_uri;
    } # managed by Certbot

    listen 80;
    listen [::]:80;
    server_name your-domain.com;
    return 404; # managed by Certbot
}
```

## Security Considerations

### Data Encryption

ssh-sync-server is designed with security in mind:

- All SSH keys are encrypted by the client before being transmitted to the server
- The server never has access to your unencrypted private keys
- Authentication employs secure challenge-response mechanisms
- Communication between client and server is encrypted using TLS

### Production Recommendations

For production deployments, we recommend:

1. Using HTTPS with valid certificates
2. Setting strong database passwords
3. Restricting access to the database container
4. Regularly backing up your database
5. Keeping the server software updated

## Technical Details

ssh-sync-server is built in Go and uses PostgreSQL for data storage. The application implements a RESTful API that the ssh-sync client communicates with, and it employs JWT tokens for authentication after initial machine setup.

### Architecture

The server consists of the following components:

- Web server handling API requests
- Database for storing encrypted keys and user information
- Authentication system for managing client access

## Maintenance

### Backing Up

To back up your data, you should primarily back up the PostgreSQL database:

```bash
# Create a backup from inside the container
docker exec -t ssh-sync-db pg_dumpall -U sshsync > ssh_sync_backup.sql

# Or use a volume backup of your PostgreSQL data directory
```

### Updating

To update to a newer version:

```bash
# Pull the latest images
docker pull therealpaulgg/ssh-sync-server:latest
docker pull therealpaulgg/ssh-sync-db:latest

# Restart your containers
docker-compose down
docker-compose up -d
```

## License

ssh-sync-server is released under the [MIT License](./LICENSE.txt).
