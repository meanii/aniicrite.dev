---
title: Goodbye kitty
slug: goodbye-kitty
date: 2024-07-24T00:00:00Z
tags: tmux, terminal, Linux
status: published
summary: Why I'm parting ways with the Kitty terminal.
---
For the past couple of years, I have been using the Kitty terminal by Kovid Goyal. Initially, it seemed like a perfect fit due to its speed and overall functionality. However, recent experiences and some fundamental disagreements with the author's stance on terminal multiplexers like tmux have led me to say goodbye to Kitty.

### Persistent Bugs in Kitty

The latest version of Kitty (0.24.2) has introduced several bugs that disrupt my workflow:

- **Failure to render text on screen when the terminal size changes.**
- **Cursor randomly changing to I-beam despite being set to block.**
- **Improper rendering of new text.**

While these issues were frustrating, they weren't the ultimate dealbreakers. I was willing to wait for a new release that might fix these bugs. However, what pushed me over the edge was discovering the author's opinion on tmux and terminal multiplexers in Kitty's FAQ.

### The Importance of tmux

I heavily rely on tmux for my development setup. It has saved my work multiple times, especially when I accidentally close my terminal. Unlike Kitty, which closes without any warning, tmux keeps my session intact, preventing any loss of work. This reliability is crucial for my productivity.

### The Author's Toxic Stance

Kitty's author has been openly critical of terminal multiplexers like tmux, even suggesting users "go soak your head" if they disagree with his views. This toxic attitude towards users and their preferences is unacceptable. As developers, we need tools that support our workflow, not hinder it.

### Additional Issues with Kitty

- **Limited Scrollback:** Kitty's refusal to use the hard disk as a cache results in limited scrollback, which is a significant drawback.
- **Excessive RAM Usage:** This is another consequence of not using the hard disk for caching.
- **Refusal to Acknowledge Bugs:** The author's reluctance to admit and fix bugs, even those that are security vulnerabilities, further undermines my trust in Kitty.

### Conclusion

Given these issues, I have decided to part ways with the Kitty terminal. tmux, despite being labeled a "hack" by Kitty's author, has proven to be robust and indispensable for my workflow. It not only saves my work but also offers functionality that Kitty cannot match, such as managing remote sessions.

To anyone facing similar frustrations with Kitty, I highly recommend giving tmux a try. It's a powerful tool that, combined with a reliable terminal emulator, can significantly enhance your development experience.

Goodbye, Kitty. Hello, tmux!

---

**Read more about the reasons behind my decision in [Gavin Howard's blog](https://gavinhoward.com/2022/02/goodbye-kitty/).**
