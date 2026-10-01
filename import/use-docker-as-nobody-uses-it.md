---
title: Using Docker as a throwaway dev environment
slug: use-docker-as-nobody-uses-it
date: 2022-04-11T00:00:00Z
tags: Docker, DevOps
status: published
summary: Instead of installing nvm, Node and MongoDB on my machine, I run the whole dev stack out of Docker images and delete it when the project is done.
---
Setting up a MERN project locally used to mean installing nvm, picking a Node version, installing MongoDB, and then fighting whichever package refused to build on that version. Do it for a few projects and the machine fills up with versions that disagree with each other.

I stopped. Now the whole stack runs out of Docker images and nothing gets installed on the host.

## Install Docker

On most Linux distributions:

```bash
curl -o- https://get.docker.com | sh -x
```

## Run the Node app

From the project directory:

```bash
sudo docker run -it -v $(pwd):/srv -w /srv -p 3000:3000 node:current npm run start:dev
```

`-it` gives you an interactive terminal, `-v` mounts the current directory into the container, `-w` sets the working directory and `-p` forwards the port.

![](https://miro.medium.com/v2/resize:fit:1400/format:webp/1*jC1ETEU62n9wVQhVcEAyvw.png)

A Vite project is the same command with a different port and script:

```bash
sudo docker run -it -v $(pwd):/srv -w /srv -p 5173:5173 node:current npm run dev
```

![](https://miro.medium.com/v2/resize:fit:1400/format:webp/1*nRjpvDFAaD5yuGDW0rAr1g.png)

## Run MongoDB

```bash
sudo docker run -d -p 27017:27017 --name my-demo-mongo mongo
```

No Node on the host, no local Mongo, nothing to uninstall later. When the project is done I delete the containers and the machine is as clean as it was before.

![](https://miro.medium.com/v2/resize:fit:1400/format:webp/1*16QMx1_smA-yr9DRyRgzVg.png)
