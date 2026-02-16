# Deploying Bank of Dad to Railway

This guide walks through deploying the app to Railway with three services:

- **Frontend** — Static React site (built by Vite, served by Railway)
- **Backend** — Go API server (Docker)
- **PostgreSQL** — Railway-managed database

## Architecture

```
                          ┌──────────────┐
    Browser ─────────────►│   Frontend   │  (static site)
       │                  │  Railway CDN │
       │                  └──────────────┘
       │
       │  API calls
       │  (https://backend-xxx.up.railway.app/api/...)
       │
       ▼
 ┌───────────┐         ┌────────────┐
 │  Backend   │────────►│ PostgreSQL │
 │  (Docker)  │         │ (Railway)  │
 │  :8001     │         │ :5432      │
 └───────────┘         └────────────┘
```

The frontend calls the backend directly via its public URL (no reverse proxy). CORS on the backend allows requests from the frontend's domain.

---

## Prerequisites

- A [Railway](https://railway.app) account
- [Railway CLI](https://docs.railway.com/guides/cli) installed: `npm install -g @railway/cli`
- Google OAuth credentials (client ID + secret) from [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
- A base64-encoded JWT secret (minimum 32 bytes)

Generate a JWT secret if you don't have one:

```bash
openssl rand -base64 64
```

---

## Step 1: Create a Railway Project

```bash
railway login
railway init
```

Choose **"Empty Project"** when prompted. Note the project name.

---

## Step 2: Provision PostgreSQL

In the Railway dashboard:

1. Open your project
2. Click **"+ New"** → **"Database"** → **"Add PostgreSQL"**
3. Railway creates the instance and exposes connection variables automatically

The key variable is `DATABASE_URL`, which you'll reference from the backend service.

---

## Step 3: Deploy the Backend

### 3a. Create the backend service

In the Railway dashboard:

1. Click **"+ New"** → **"GitHub Repo"** → select **bank-of-dad**
2. Railway detects the repo. Before deploying, configure the service:

In the service **Settings** tab:

| Setting | Value |
|---------|-------|
| **Root Directory** | `backend` |
| **Builder** | Dockerfile |

Railway will find `backend/Dockerfile` automatically.

### 3b. Set environment variables

In the backend service's **Variables** tab, add:

| Variable | Value | Notes |
|----------|-------|-------|
| `DATABASE_URL` | `${{Postgres.DATABASE_URL}}` | Railway variable reference — auto-populated from the Postgres service |
| `GOOGLE_CLIENT_ID` | `your-google-client-id` | From Google Cloud Console |
| `GOOGLE_CLIENT_SECRET` | `your-google-client-secret` | From Google Cloud Console |
| `JWT_SECRET` | `your-base64-secret` | The base64 string from `openssl rand -base64 64` |
| `FRONTEND_URL` | `https://your-frontend-domain.up.railway.app` | Set after the frontend is deployed (Step 4). Come back and update this. |
| `GOOGLE_REDIRECT_URL` | `https://your-backend-domain.up.railway.app/api/auth/google/callback` | Use the backend service's Railway-assigned domain |
| `SERVER_PORT` | `8001` | Matches the `EXPOSE` in the Dockerfile |

> **Note:** `DATABASE_URL` uses Railway's [reference variable syntax](https://docs.railway.com/guides/variables#reference-variables) (`${{Postgres.DATABASE_URL}}`), which auto-resolves to the internal connection string. If the Postgres service has a different name in your project, adjust accordingly.

### 3c. Expose the backend publicly

The backend needs a public URL so the frontend can call it:

1. Go to the backend service's **Settings** → **Networking**
2. Click **"Generate Domain"** to get a `*.up.railway.app` domain
3. Note this URL — you'll need it for `VITE_API_URL`, `GOOGLE_REDIRECT_URL`, and Google OAuth settings

### 3d. Deploy

Railway auto-deploys on push to your connected branch. To trigger manually:

```bash
railway up --service backend
```

Database migrations run automatically on startup — no manual migration step needed.

---

## Step 4: Deploy the Frontend

### 4a. Create the frontend service

In the Railway dashboard:

1. Click **"+ New"** → **"GitHub Repo"** → select **bank-of-dad** (same repo, second service)
2. Configure the service:

In the service **Settings** tab:

| Setting | Value |
|---------|-------|
| **Root Directory** | `frontend` |
| **Builder** | Nixpacks |
| **Build Command** | `npm run build` |
| **Start Command** | `npx serve dist -s -l 8000` |

> Railway's Nixpacks builder detects the Node project, installs dependencies, runs the build, and serves the static output. The `-s` flag enables SPA fallback (rewrites all routes to `index.html`). Install `serve` as a dependency if not already present: add it to `package.json` devDependencies or use `npx`.

**Alternative:** If you prefer not to use `serve`, you can keep the Nginx Docker approach. Set the builder to **Dockerfile** and update `nginx.conf` to remove the `/api/` proxy block (since API calls go directly to the backend).

### 4b. Set environment variables

| Variable | Value | Notes |
|----------|-------|-------|
| `VITE_API_URL` | `https://your-backend-domain.up.railway.app` | The backend's public Railway domain (no trailing slash). Must be prefixed with `VITE_` for Vite to embed it at build time. |

### 4c. Generate a public domain

1. Go to the frontend service **Settings** → **Networking**
2. Click **"Generate Domain"**
3. Note this URL — go back and set it as the backend's `FRONTEND_URL` variable (Step 3b)

### 4d. Deploy

```bash
railway up --service frontend
```

---

## Step 5: Configure Google OAuth

After both services are deployed and have their Railway domains:

1. Go to [Google Cloud Console → Credentials](https://console.cloud.google.com/apis/credentials)
2. Edit your OAuth 2.0 Client ID
3. Under **Authorized redirect URIs**, add:
   ```
   https://your-backend-domain.up.railway.app/api/auth/google/callback
   ```
4. Under **Authorized JavaScript origins**, add:
   ```
   https://your-frontend-domain.up.railway.app
   ```
5. Save

---

## Step 6: Update Cross-References

Now that both services have their domains, make sure these variables are set correctly:

| Service | Variable | Value |
|---------|----------|-------|
| Backend | `FRONTEND_URL` | `https://your-frontend-domain.up.railway.app` |
| Backend | `GOOGLE_REDIRECT_URL` | `https://your-backend-domain.up.railway.app/api/auth/google/callback` |
| Frontend | `VITE_API_URL` | `https://your-backend-domain.up.railway.app` |

> Changing `VITE_API_URL` requires a **rebuild** of the frontend since Vite embeds it at build time. Railway will auto-redeploy when you change the variable.

---

## Step 7: Custom Domains (Optional)

To use a custom domain (e.g., `bankofdad.com`):

1. In the frontend service **Settings** → **Networking** → **Custom Domain**, enter your domain
2. Railway provides DNS records (CNAME) — add them at your registrar
3. Update the backend's `FRONTEND_URL` to match the custom domain
4. Update `GOOGLE_REDIRECT_URL` and Google Console redirect URIs if the backend also gets a custom domain
5. Rebuild the frontend with the updated `VITE_API_URL` if the backend domain changes

---

## Verification Checklist

After deployment:

- [ ] Backend health: visit `https://your-backend-domain.up.railway.app/api/health` (or any known endpoint)
- [ ] Frontend loads: visit `https://your-frontend-domain.up.railway.app`
- [ ] Google OAuth login works end-to-end (redirects to Google → back to app → lands on dashboard or setup)
- [ ] Child login works at `https://your-frontend-domain.up.railway.app/<family-slug>`
- [ ] API calls work (balances load, transactions display)
- [ ] Check Railway logs for both services if anything fails

---

## Environment Variable Reference

### Backend

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | `postgres://bankofdad:bankofdad@localhost:5432/bankofdad?sslmode=disable` | PostgreSQL connection string |
| `GOOGLE_CLIENT_ID` | Yes | — | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | Yes | — | Google OAuth client secret |
| `JWT_SECRET` | Yes | — | Base64-encoded key, minimum 32 bytes decoded |
| `GOOGLE_REDIRECT_URL` | No | `http://localhost:{port}/api/auth/google/callback` | OAuth callback URL |
| `SERVER_PORT` | No | `8001` | Port the backend listens on |
| `FRONTEND_URL` | No | `http://localhost:8000` | Frontend origin for CORS and OAuth redirects |

### Frontend

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VITE_API_URL` | Yes (prod) | `''` (empty, uses relative paths) | Backend's full public URL. Must start with `VITE_` for Vite to embed at build time. |

---

## Troubleshooting

**CORS errors in browser console**
→ Verify `FRONTEND_URL` on the backend matches the exact frontend domain (including `https://`, no trailing slash).

**OAuth redirect fails or "Invalid state parameter"**
→ Verify `GOOGLE_REDIRECT_URL` on the backend points to the backend's own domain (`/api/auth/google/callback`), and that the same URL is listed in Google Cloud Console's authorized redirect URIs.

**API calls return 404**
→ Check `VITE_API_URL` is set correctly on the frontend service. Remember that changing it requires a rebuild (Railway does this automatically when variables change).

**Database connection refused**
→ Ensure `DATABASE_URL` uses the Railway reference variable `${{Postgres.DATABASE_URL}}` so it resolves to the internal connection string.

**"JWT_SECRET must be at least 32 bytes"**
→ The value must be base64-encoded. `openssl rand -base64 64` produces a valid 64-byte key.
