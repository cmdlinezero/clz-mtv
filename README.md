# mtv (Markdown to Video/Cast)

![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?style=flat-square&logo=go)
![Tag](https://img.shields.io/github/v/tag/cmdlinezero/clz-lhd?label=Tag)
![License](https://img.shields.io/badge/license-GNU-green?style=flat-square)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen?style=flat-square)
![Asciinema](https://img.shields.io/badge/platform-asciinema-red?style=flat-square)

`mtv` is a powerful CLI tool designed to transform Markdown documents into professional-grade terminal recordings (`.cast` files). It automates the process of "typing" commands and displaying syntax-highlighted outputs, making it perfect for creating documentation demos, tutorials, and lab walkthroughs.

<img src="https://github.com/cmdlinezero/clz-mtv/blob/main/screenshots/mtv.gif" width="600">

## Features

* **Automated Typing**: Simulates natural human typing speeds with randomized delays.
* **Markdown Driven**: Convert standard fenced code blocks directly into terminal sequences.
* **Syntax Highlighting**: Automatically applies ANSI color highlighting to command outputs using [Chroma](https://github.com/alecthomas/chroma).
* **Special Directives**: Use postfixes like `info`, `warn`, and `output` to control terminal behavior.
* **Deterministic IDs**: Generates SHA-1 based IDs for command tracking.
* **Asciinema Compatible**: Produces `.cast` files (v2) playable via `asciinema` or web players.

## Installation

Ensure you have Go installed, then clone the repository and build the binary:

```bash
go build -o mtv main.go
```

## Usage

`mtv` operates in two primary stages: **Convert** (Markdown → JSON) and **Generate** (JSON → .cast).

### 1. Convert Markdown to JSON
This step parses your Markdown and prepares the commands and responses.

```bash
./mtv convert -i instructions.md -o commands.json --theme monokai
```

### 2. Generate the .cast File
This step "records" the session based on the JSON schema.

```bash
./mtv generate -i commands.json -o demo.cast --prompt "rosera@labdemo.app:~$ "
```

### 3. Create an asciinema video for the demo
This step plays the generated cast file using the asciinema utiity.

```bash
asciinema play demo.cast
```

### 4. (OPTIONAL) Generate a GIF

This step is optional and is used to generate a GIF for the demo.

```bash
agg --font-size 32 demo.cast demo.gif
```

## Markdown Syntax Guide

`mtv` uses the **postfix** of a fenced code block to determine how to handle content.

### Normal Commands
Standard code blocks are treated as commands to be typed into the terminal.

````markdown
```bash
ls -la
mkdir new_project
```
````

### Syntax Highlighted Output
Use the `output` postfix (or other file types like `yaml`, `json`, `dockerfile`) to attach a syntax-highlighted response to the *previous* command.

````markdown
```bash
cat config.yaml
```

```yaml output
version: '3'
services:
  api:
    image: node:latest
```
````

### Messages (Info & Warn)
Use `info` or `warn` to display commented messages in the terminal. These are great for explaining steps without executing them.

````markdown
```text info
We are now going to initialize the database.
This might take a few seconds.
```
````

## CLI Options

### Global Flags
* `-i, --input`: Input file path (Required).
* `-o, --output`: Output file path.

### Convert Flags
* `-t, --theme`: The Chroma syntax highlighting theme (default: `monokai`).

### Generate Flags
* `-p, --prompt`: The terminal prompt string (default: `rosera@labdemo.app:~$ `).
* `--width`: Terminal width (default: `100`).
* `--height`: Terminal height (default: `30`).

## Playing the Output
The resulting `.cast` file can be played back in your terminal:

```bash
asciinema play demo.cast
```

## License
GNU V3
