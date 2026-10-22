<h1 align="center">
  <img alt="autobrr logo" src=".github/images/logo.png" width="160px"/><br/>
  autobrr
</h1>

<p align="center">autobrr redefines download automation for torrents, drawing inspiration from tools like trackarr, autodl-irssi, and flexget. We've combined the best of these worlds into one versatile tool that can do it all, and then some.</p>

<p align="center"><img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/autobrr/autobrr?style=for-the-badge">&nbsp;<img alt="GitHub all releases" src="https://img.shields.io/github/downloads/autobrr/autobrr/total?style=for-the-badge">&nbsp;<img alt="GitHub Workflow Status" src="https://img.shields.io/github/actions/workflow/status/autobrr/autobrr/release.yml?style=for-the-badge"></p>

<img alt="autobrr ui" src=".github/images/front-dark.png" /><br/>
*We also have a light theme. Check it out [here](.github/images/front-light.png).*

## Documentation

Full documentation can be found at [https://autobrr.com](https://autobrr.com)

## Table of Contents

1. [What Is Autobrr?](#what-is-autobrr)
2. [Key Features](#key-features)
   - [Available Actions](#available-actions)
   - [RSS Support](#rss-support)
3. [Installation](#installation)
4. [Community](#community)
5. [Contributing](#contributing)
6. [Code of Conduct](#code-of-conduct)
7. [License](#license)

## What Is Autobrr?

What is autobrr, and how does it fit into the ecosystem?

We can start by talking about torrent trackers (hereby referred to as indexers) and maintaining ratio. You are required
to maintain a ratio with most indexers. Ratio is built by seeding your torrents. The earlier you're seeding a torrent,
the more peers you make yourself available to on that torrent.

Many tools use RSS to look for new torrents. RSS feeds are updated regularly, but too slow to
let you be a part of what we call the initial swarm of a torrent. This is where autobrr comes into play.

Many indexers announce new torrents on their IRC channels the second it is uploaded to the site. autobrr monitors such
channels in real time and grabs the torrent file as soon as it's uploaded based on certain conditions (hereby referred
to as filters) that you set up within autobrr. It then hands that torrent file on via an action set within the filter:
autobrr can drop matches (torrent files that meet your filter's criteria) into a watch folder or call a webhook.

## Key Features

- Support for 100+ torrent trackers with IRC announces
- Torznab and RSS support to easily get access to hundreds of torrent indexers
- Torrent Magnet support
- Powerful but simple filtering with RegEx support (like in autodl-irssi)
- Native iOS app (with dark mode!) to manage everything
- Runs entirely on your iPhone or iPad, with no server to host
- SQLite database
- Notifications (built-in inbox and webhooks)

### Available Actions

- **Watch Folder:** Monitors specified folders for new files.
- **Webhook:** Offers webhook support for diverse integration needs.

### RSS Support

A lot of indexers do not announce new torrents in an IRC channel. You can still make use of these indexers with autobrr
since it has built-in support for feeds as well. Both Torznab and regular RSS are supported. RSS indexers are treated the same way as regular indexers within autobrr.

While there are a lot of applications that handle RSS well, we think autobrr offers very easy-to-use filtering to help
you get the content you want.

## Installation

autobrr runs as a standalone iOS app that you build and sideload yourself. The engine runs inside the app, so there is no server to install. See [ios/README.md](ios/README.md) for build and sideloading steps.

Head over to our [Configuration Guide](https://autobrr.com/configuration/autobrr) to learn how to set up your indexers, IRC, and filters.

## Community

Join our friendly and welcoming community on [Discord](https://discord.autobrr.com)! Connect with fellow autobrr users, get advice, and share your experiences. Whether you're seeking help, wanting to contribute, or just looking to discuss your ideas, our community is a hub of discussion and support. We're all here to help each other out, so don't hesitate to jump in!

## Contributing

Whether you're fixing a bug, adding a feature, or improving documentation, your help is appreciated. Here's how you can contribute:

### Reporting Issues and Suggestions

- **Report Bugs:** Encountered a bug? Please report it using our [bug report template](/.github/ISSUE_TEMPLATE/bug_report.md). Include detailed steps to reproduce, expected behavior, and any relevant screenshots or logs.
- **Feature Requests:** Submit your feature request by [opening a new idea in our discussions](https://github.com/autobrr/autobrr/discussions/new?category=ideas). Describe your idea and how it will improve `autobrr`.

### Code Contributions

Check out the full guide for contributing [here](CONTRIBUTING.md).

- **Fork and Clone:** Fork the `autobrr` repository and clone it to start working on your changes.
- **Branching:** Create a new branch for your changes. Use a descriptive name for easy understanding.
- **Coding:** Ensure your code is well-commented for clarity.
- **Commit Guidelines:** We appreciate the use of [Conventional Commit Guidelines](https://www.conventionalcommits.org/en/v1.0.0/#summary) when writing your commits.
  - There is no need for force pushing or rebasing. We squash commits on merge to keep the history clean and manageable.
- **Pull Requests:** Submit a pull request with a clear description of your changes. Reference any related issues.
- **Code Review:** Be open to feedback during the code review process.

### Documentation

See an area that needs clarity or additional information? Feel free to update our documentation [here](https://github.com/autobrr/autobrr.com).

## Code of Conduct

We follow a code of conduct that promotes respectful and harassment-free experiences. Please read [our Code of Conduct](CODE_OF_CONDUCT.md) before participating.

## License

autobrr is proudly open-source and is released under the [GNU General Public License v2 or later (GPLv2+)](https://www.gnu.org/licenses/old-licenses/gpl-2.0-standalone.html). This license allows you the freedom to run, study, share, and modify the software:

- **Freedom to Run:** You can run autobrr in any environment, for any purpose.
- **Freedom to Study and Modify:** Access to the source code allows you to study and modify autobrr to suit your needs.
- **Freedom to Share:** You can redistribute copies of autobrr to help others.
- **Freedom to Enhance:** Contributions to improve autobrr are always welcome.

Copyright 2021-2024
