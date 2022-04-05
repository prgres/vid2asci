# VID2ASCI


## Usage

```
❯❯❯ go run .
NAME:
   vid2asci - A new cli application

USAGE:
   vid2asci [global options] command [command options] [arguments...]

DESCRIPTION:
   Run your favorite shows directly in your favorite term (with some laggy and fuzzy frames of course

AUTHOR:
   M. Więcek

COMMANDS:
   render
   play
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug     debug (default: false)
   --help, -h  show help (default: false)
```

### Render

Render input file <`./video.mp4`> to asci frames under the `./asci` folder.
```
❯❯❯ go run . --debug render -i ./video.mp4
```

### Play

Play video in term.

```
❯❯❯ go run . --debug play
```
