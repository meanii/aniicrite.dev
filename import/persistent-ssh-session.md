---
title: autossh — Persistent your SSH sessions even if drops
slug: persistent-ssh-session
date: 2023-11-04T00:00:00Z
tags: SSH, autossh, Linux
status: published
summary: Keep your SSH sessions alive across network drops with autossh.
---
![ssh-session](https://miro.medium.com/v2/resize:fit:1400/format:webp/1*iVYccCsXIgpYvmC3zMnLBA.png)

In the world of remote connectivity, maintaining an uninterrupted SSH session is crucial. Unforeseen network issues can often lead to dropped connections, disrupting work processes and causing inconvenience. To mitigate this, a handy tool called assh (autossh) comes to the rescue. Developed to assist in automatically reconnecting an SSH connection if it happens to drop, assh ensures a seamless and persistent SSH experience, even in the face of intermittent network problems.

[Watch the demo of autossh](https://www.youtube.com/watch?v=EAjosu4AVGQ)

## Installation

You can install assh by running the following command in your terminal:

```bash
curl --silent -o- https://raw.githubusercontent.com/meanii/assh/main/install.sh | sudo bash
```

**Note**: this will require sudo privileges, as it will install the script to `/usr/local/bin/assh`

## Usage

```bash
assh <ssh-connection-string>
```

github: <https://github.com/meanii/assh>
