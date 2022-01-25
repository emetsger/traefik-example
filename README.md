# Traefik Proxy Demo

This demo will listen to port 80 and port 8080, but you can adjust the values of `TRAEFIK_PROXY_PORT` and `TRAEFIK_CONSOLE_PORT` to different values.

After checking out the code repository:
1. Build the images: `docker-compose build`
2. Launch containers: `docker-compose up -d` (or if you prefer to see the logs, leave off `-d`)

The Traefik console is available by default at [http://localhost:8080], and proxied services are behind port 80.

This demo is configured to use hostnames from the [`traefik.me`](http://traefik.me) domain, which returns `127.0.0.1` for all lookups.

Visit [http://server-01.traefik.me/hello-world.html], which is a proxied service that is automatically configured using metadata in the [docker-compose.yml](./docker-compose.yml) file. 