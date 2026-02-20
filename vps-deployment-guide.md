# Deploying DML API to VPS

Since your VPS already has Docker containers running, we built this new API stack to be completely isolated. It binds to local port `8081` internally and does **not** compete for ports `80` or `443`.

This means your Nginx server (which is already listening on `80`/`443`) just needs a new server block to route traffic for `api.inova.krd` directly into our new container!

## 1. Setup the Files
1. SSH into the VPS.
2. Clone the repository to `/opt/dml-api` (or your preferred folder).
3. Set up the `.env` file:
   ```bash
   cp .env.example .env
   nano .env
   ```
4. Define your `JWT_SECRET` and `DB_PASSWORD`.
5. Start the isolated stack:
   ```bash
   docker compose up -d
   ```

## 2. Nginx Reverse Proxy Configuration
You need to tell Nginx to listen for `api.inova.krd` and forward it to the `8081` port.

1. Create a new Nginx site configuration:
   ```bash
   sudo nano /etc/nginx/sites-available/api.inova.krd
   ```

2. Paste the following configuration exactly:
   ```nginx
   server {
       listen 80;
       server_name api.inova.krd;

       location / {
           # Proxy to our isolated Docker API port
           proxy_pass http://localhost:8081;
           
           # Preserve standard client mapping headers
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
           proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
           proxy_set_header X-Forwarded-Proto $scheme;
           
           # Allow CORS preflight caching natively
           add_header 'Access-Control-Max-Age' 1728000;
       }
   }
   ```

3. Enable the site and restart Nginx:
   ```bash
   sudo ln -s /etc/nginx/sites-available/api.inova.krd /etc/nginx/sites-enabled/
   sudo nginx -t
   sudo systemctl restart nginx
   ```

## 3. Secure with HTTPS (Let's Encrypt)
Run Certbot to automatically provision the SSL certificate and upgrade your Nginx block to HTTPS natively:
```bash
sudo certbot --nginx -d api.inova.krd
```

*(Certbot will automatically edit the Nginx file you just made to inject the standard port 443 SSL certificates).*
