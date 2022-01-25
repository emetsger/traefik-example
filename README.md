# Traefik Proxy Demo

This demo will listen to port 80 and port 8080, but you can adjust the values of `TRAEFIK_PROXY_PORT` and `TRAEFIK_CONSOLE_PORT` to different values.

After checking out the code repository:
1. Build the images: `docker-compose build`
2. Launch containers: `docker-compose up` (it may be useful to view the logs during the sample exercises)

The Traefik console is available by default at http://localhost:8080, and proxied services are behind port 80.

This demo is configured to use hostnames from the [`traefik.me`](http://traefik.me) domain, which returns `127.0.0.1` for all DNS lookups.

# Exercises

These exercises demonstrate some of Traefik's basic capabilities.  Note that the developer makes the decision about how requests are routed to containers by specifying [routes](https://doc.traefik.io/traefik/routing/routers/) in the [provider's (i.e. Docker) metadata](https://doc.traefik.io/traefik/providers/docker/). Any tool that is able to label a Docker image can be used to provide this information, or it can be embedded in the image by the developer.

1. Navigate to the Traefik Console, and view the [services](http://localhost:8080/dashboard/#/http/services).  Note the presence of the example services. Traefik discovered them by using the labels configured in `docker-compose.yml`.
2. Visit http://service.traefik.me/index.html.  Traefik knows how to route the request based on the HTTP host header, as specified by the rule `Host\(\`service.traefik.me\`\)`.  Note the `group` in the HTTP response; it's simply a key that all containers started from the same service will share.
3. Visit http://service.traefik.me/group2/index.html.  Note that the group has changed.  Traefik routes this request based on the HTTP host header and the path prefix: `Host\(\`service.traefik.me\`\) && PathPrefix\(\`/group2/\`\)`
4. Go ahead and stop one of the services, e.g. `docker-compose down server`, and see how traefik responds.
5. Now launch multiple instances of a service by invoking `docker-compose up --scale server=5`, then re-execute your request to http://service.traefik.me/index.html multiple times.  You should notice that traefik automatically discovered the new instances, and load balances across them automatically.