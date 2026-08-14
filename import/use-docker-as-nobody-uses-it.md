---
title: Using Docker as a throwaway dev environment
slug: use-docker-as-nobody-uses-it
date: 2022-04-11T00:00:00Z
tags: Docker, DevOps
status: published
summary: Skip installing nvm and databases locally — run your dev stack straight from Docker images.
---
Setting up a MERN stack locally usually means installing nvm, picking a Node version, installing MongoDB, and then juggling versions when a package won't build. It piles up fast and the versions start fighting each other.

I stopped doing that. Instead I run the whole thing out of Docker images and keep my machine clean.

### Install Docker

Works on most Linux distros:

```bash
curl -o- https://get.docker.com | sh -x
```

### Run a Node app

From your project directory:

```bash
sudo docker run -it -v $(pwd):/srv -w /srv -p 3000:3000 node:current npm run start:dev
```

- `-it` — interactive terminal
- `-v` — mount the current dir into the container
- `-w` — set the working directory
- `-p` — forward the port

![](https://miro.medium.com/v2/resize:fit:1400/format:webp/1*jC1ETEU62n9wVQhVcEAyvw.png)

Same idea for a Vite React project:

```bash
sudo docker run -it -v $(pwd):/srv -w /srv -p 5173:5173 node:current npm run dev
```

![](https://miro.medium.com/v2/resize:fit:1400/format:webp/1*nRjpvDFAaD5yuGDW0rAr1g.png)

### Run MongoDB

```bash
sudo docker run -d -p 27017:27017 --name my-demo-mongo mongo
```

No Node versions on my host, no local Mongo, nothing to uninstall later. When I'm done with a project I delete the containers and that's it.

![](https://miro.medium.com/v2/resize:fit:1400/format:webp/1*16QMx1_smA-yr9DRyRgzVg.png)
