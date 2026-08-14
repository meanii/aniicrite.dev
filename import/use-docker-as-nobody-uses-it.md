---
title: Use Docker Like Nobody Else 💫
slug: use-docker-as-nobody-uses-it
date: 2022-04-11T00:00:00Z
tags: Docker, DevOps
status: published
summary: Learn to use Docker to simplify and isolate your development environment.
---
Setting up a development environment can be a daunting task. Imagine you need to work with the MERN Stack. Here's what you typically need to do:

1. Install nvm or any Node.js version.
2. Install the MongoDB server.
3. If a package doesn't support your current Node.js version, you have to install a new one.
4. Keeping all these running on your local machine can be cumbersome and may cause conflicts.

So, how can we make this process easier to manage and isolate? The answer is Docker. Let's dive in!

### Installing Docker

Here's a simple script to install Docker on any Linux distro:

```bash
curl -o- https://get.docker.com | sh -x
```

### Running a Node.js Application with Docker

First, navigate to your working directory and run the following command:

```bash
sudo docker run -it -v $(pwd):/srv -w /srv -p 3000:3000 node:current npm run start:dev
```

- `-it` : interactive terminal
- `-v`: volume mount
- `-w`: working directory
- `-p`: port forwarding

![](https://miro.medium.com/v2/resize:fit:1400/format:webp/1*jC1ETEU62n9wVQhVcEAyvw.png)

This command runs a Vite React.js project:

```bash
sudo docker run -it -v $(pwd):/srv -w /srv -p 5173:5173 node:current npm run dev
```

![](https://miro.medium.com/v2/resize:fit:1400/format:webp/1*nRjpvDFAaD5yuGDW0rAr1g.png)

### Running MongoDB with Docker

```bash
sudo docker run -d -p 27017:27017 --name my-demo-mongo mongo
```

By using Docker, you can easily manage and isolate your development environment, avoiding conflicts and making the setup process much smoother.

![](https://miro.medium.com/v2/resize:fit:1400/format:webp/1*16QMx1_smA-yr9DRyRgzVg.png)
