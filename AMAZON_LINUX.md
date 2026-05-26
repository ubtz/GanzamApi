# Run GanzamApi on Amazon Linux

## Build

From Windows PowerShell:

```powershell
.\scripts\build-amazon-linux.ps1
```

From Linux/macOS:

```bash
./scripts/build-amazon-linux.sh
```

The Linux binary is created at `dist/GanzamApi`.

## Copy files to the server

Copy these into `/opt/ganzamapi`:

```text
dist/GanzamApi
conf/
static/
views/
```

Copy `deploy/ganzamapi.env.example` to `/etc/ganzamapi/ganzamapi.env` and fill in real values. On Amazon Linux, the important values are:

```text
APP_ENV=prod
DB_SERVER=your reachable SQL Server host
DB_PORT=1433
DB_USER=your database user
DB_PASSWORD=your database password
DB_NAME=Ganzam
JWT_SECRET=replace with a long random value
AWS_REGION=ap-northeast-1
AWS_S3_BUCKET=your bucket
```

## Install as a service

On the Amazon Linux server:

```bash
sudo useradd --system --home /opt/ganzamapi --shell /sbin/nologin ganzamapi || true
sudo mkdir -p /opt/ganzamapi /etc/ganzamapi
sudo chown -R ganzamapi:ganzamapi /opt/ganzamapi
sudo chmod +x /opt/ganzamapi/GanzamApi
sudo chmod 600 /etc/ganzamapi/ganzamapi.env
sudo cp deploy/ganzamapi.service /etc/systemd/system/ganzamapi.service
sudo systemctl daemon-reload
sudo systemctl enable --now ganzamapi
sudo systemctl status ganzamapi
```

The app listens on `HTTP_ADDR` and `HTTP_PORT`. The deployment template uses `0.0.0.0:8080`.

## Logs

```bash
sudo journalctl -u ganzamapi -f
```

## Health check

```bash
curl http://127.0.0.1:8080/version
```

If you see the app trying to connect to `172.30.30.30`, then `APP_ENV` or the `DB_*` variables are not being loaded by the running service:

```bash
sudo systemctl show ganzamapi --property=EnvironmentFiles
sudo journalctl -u ganzamapi -n 100 --no-pager
```
